package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"corridor/internal/storage"
)

// IngestStats records summary metrics from ingesting national park highway corridors.
type IngestStats struct {
	WaysParsed      int
	SegmentsCreated int
	TotalLengthM    float64
	Highways        []string
}

type osmElement struct {
	Type     string            `json:"type"`
	ID       int64             `json:"id"`
	Lat      *float64          `json:"lat,omitempty"`
	Lon      *float64          `json:"lon,omitempty"`
	Nodes    []int64           `json:"nodes,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
	Geometry []osmPoint        `json:"geometry,omitempty"`
}

type osmPoint struct {
	Lon *float64 `json:"lon"`
	Lat *float64 `json:"lat"`
}

type osmDocument struct {
	Remark   string       `json:"remark"`
	Elements []osmElement `json:"elements"`
}

var (
	re191Ref  = regexp.MustCompile(`(?i)\b(US[- ]?191|U\.S\.[- ]?191|^191$|;191|191;)\b`)
	re89Ref   = regexp.MustCompile(`(?i)\b(US[- ]?89|U\.S\.[- ]?89|^89$|;89|89;)\b`)
	re191Name = regexp.MustCompile(`(?i)\b(US[- ]?191|U\.S\.[- ]?191|Highway 191|Hwy 191)\b`)
	re89Name  = regexp.MustCompile(`(?i)\b(US[- ]?89|U\.S\.[- ]?89|Highway 89|Hwy 89)\b`)
)

// NPSRoadArtifact returns the provenance artifact metadata for Yellowstone & Grand Teton highways.
func NPSRoadArtifact() Artifact {
	return Artifact{
		SourceID:    "f1e948c2-3e4b-4b19-9f4a-8d62688f4b10",
		Name:        "OpenStreetMap · Yellowstone & Grand Teton Highways (US-191 / US-89)",
		URL:         "https://www.openstreetmap.org/copyright",
		License:     "ODbL 1.0 · © OpenStreetMap contributors",
		SHA256:      "",
		ObjectKey:   "",
		RetrievedAt: time.Date(2026, 10, 11, 14, 0, 0, 0, time.UTC),
	}
}

// IngestNPSCorridors ingests OSM highway corridors for US-191 and US-89 in the Greater Yellowstone
// and Grand Teton national parks ecosystem, discretizes ways into 0.1-mile segments using
// PostGIS discretize_highway, and records them in road_segments.
func IngestNPSCorridors(ctx context.Context, pool *pgxpool.Pool, objects *storage.Client, path string) (IngestStats, error) {
	if ctx == nil {
		return IngestStats{}, errors.New("context is required")
	}
	if pool == nil {
		return IngestStats{}, errors.New("database pool is required")
	}
	if path == "" {
		return IngestStats{}, errors.New("path is required")
	}

	// #nosec G304 -- operator path in production; test fixtures in testing.
	file, err := os.Open(path)
	if err != nil {
		return IngestStats{}, fmt.Errorf("open nps highway source: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("resource cleanup failed", "error", err)
		}
	}()

	hash := sha256.New()
	stat, err := file.Stat()
	if err != nil {
		return IngestStats{}, fmt.Errorf("stat file: %w", err)
	}

	if _, err := io.Copy(hash, file); err != nil {
		return IngestStats{}, fmt.Errorf("hash nps highway source: %w", err)
	}
	shaHex := hex.EncodeToString(hash.Sum(nil))

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return IngestStats{}, fmt.Errorf("rewind file: %w", err)
	}

	artifact := NPSRoadArtifact()
	artifact.SHA256 = shaHex
	artifact.ObjectKey = "sha256/" + shaHex + "/yellowstone-teton-highways.json"

	if objects != nil {
		if err := objects.Archive(ctx, artifact.ObjectKey, file, stat.Size()); err != nil {
			return IngestStats{}, fmt.Errorf("archive nps highways: %w", err)
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return IngestStats{}, fmt.Errorf("rewind file after archive: %w", err)
		}
	}

	var doc osmDocument
	decoder := json.NewDecoder(io.LimitReader(file, 64<<20))
	if err := decoder.Decode(&doc); err != nil {
		return IngestStats{}, fmt.Errorf("decode osm json: %w", err)
	}
	if len(doc.Elements) == 0 {
		return IngestStats{}, errors.New("empty or invalid osm document")
	}

	// 1. First pass: index milestone reference nodes
	refNodes := make(map[int64]float64)
	refNodeCoords := make(map[int64][2]float64)

	for _, el := range doc.Elements {
		if el.Type != "node" || el.Lat == nil || el.Lon == nil {
			continue
		}
		if err := validateWGS84(*el.Lon, *el.Lat); err != nil {
			return IngestStats{}, err
		}
		if mp, ok := extractNodeMilepost(el.Tags); ok {
			refNodes[el.ID] = mp
			refNodeCoords[el.ID] = [2]float64{*el.Lon, *el.Lat}
		}
	}

	// 2. Second pass: filter and validate highway ways for US-191 and US-89
	type parsedWay struct {
		way           osmElement
		highwayRef    string
		hasExplicitMP bool
		startMP       float64
		points        [][2]float64
		state         string
	}

	var highwayWays []parsedWay
	seenHighways := make(map[string]bool)

	for _, el := range doc.Elements {
		if el.Type != "way" {
			continue
		}
		highwayRef, isTarget := matchHighwayRef(el.Tags)
		if !isTarget {
			continue
		}
		if len(el.Geometry) < 2 {
			continue
		}

		// Validate coordinates
		points := make([][2]float64, 0, len(el.Geometry))
		outsideBbox := false
		for _, p := range el.Geometry {
			if p.Lon == nil || p.Lat == nil {
				return IngestStats{}, errors.New("nil coordinate in way geometry")
			}
			if err := validateWGS84(*p.Lon, *p.Lat); err != nil {
				return IngestStats{}, err
			}
			if !inYellowstoneCorridorBbox(*p.Lon, *p.Lat) {
				outsideBbox = true
				break
			}
			points = append(points, [2]float64{*p.Lon, *p.Lat})
		}
		if outsideBbox || len(points) < 2 {
			// Skip ways outside Yellowstone / Grand Teton corridor bounds
			continue
		}

		// Compute start milepost from explicit way tags or milestone at start node
		startMP, hasMP := extractWayMilepost(el.Tags)
		if !hasMP && len(el.Nodes) > 0 {
			// Match startMP from milestone nodes against the starting node (el.Nodes[0])
			if mp, ok := refNodes[el.Nodes[0]]; ok {
				startMP = mp
				hasMP = true
			}
		}
		if !hasMP && len(points) > 0 {
			// Check if any reference node is near the start point (< 150m / ~0.0015 deg)
			startLon, startLat := points[0][0], points[0][1]
			for nID, coords := range refNodeCoords {
				dLon := coords[0] - startLon
				dLat := coords[1] - startLat
				if (dLon*dLon + dLat*dLat) < 0.000004 { // ~200m
					startMP = refNodes[nID]
					hasMP = true
					break
				}
			}
		}

		state := determineState(el.Tags, points[0][1], points[0][0])
		highwayWays = append(highwayWays, parsedWay{
			way:           el,
			highwayRef:    highwayRef,
			hasExplicitMP: hasMP,
			startMP:       startMP,
			points:        points,
			state:         state,
		})
		seenHighways[highwayRef] = true
	}

	if len(highwayWays) == 0 {
		return IngestStats{}, nil
	}

	// 3. Database transaction: register source, artifact, and slice into road_segments
	tx, err := pool.Begin(ctx)
	if err != nil {
		return IngestStats{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Warn("resource cleanup failed", "error", err)
		}
	}()

	// Advisory lock to serialize NPS road corridor ingestion
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(827369)"); err != nil {
		return IngestStats{}, fmt.Errorf("lock nps road import: %w", err)
	}

	// Register source for provenance
	if _, err = tx.Exec(ctx, `
		INSERT INTO sources(id, name, url, license, status, public_url)
		VALUES ($1, $2, $3, $4, 'approved', $3)
		ON CONFLICT(id) DO UPDATE SET
			name = EXCLUDED.name,
			license = EXCLUDED.license,
			public_url = EXCLUDED.public_url
	`, artifact.SourceID, artifact.Name, artifact.URL, artifact.License); err != nil {
		return IngestStats{}, fmt.Errorf("register source: %w", err)
	}

	// Register source artifact
	if _, err = tx.Exec(ctx, `
		INSERT INTO source_artifacts(source_id, object_key, retrieved_at, sha256)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT(source_id, sha256) DO NOTHING
	`, artifact.SourceID, artifact.ObjectKey, artifact.RetrievedAt, artifact.SHA256); err != nil {
		return IngestStats{}, fmt.Errorf("register source artifact: %w", err)
	}

	var stats IngestStats
	for h := range seenHighways {
		stats.Highways = append(stats.Highways, h)
	}

	corridorMP := make(map[string]float64)

	// Slice each way into 0.1-mile segments and insert into road_segments
	for _, hw := range highwayWays {
		geomBytes, err := json.Marshal(struct {
			Type        string       `json:"type"`
			Coordinates [][2]float64 `json:"coordinates"`
		}{
			Type:        "LineString",
			Coordinates: hw.points,
		})
		if err != nil {
			return stats, fmt.Errorf("marshal way geometry: %w", err)
		}

		// Ensure each OSM way LineString is passed individually (handling MultiLineStrings cleanly via ST_Dump)
		const sliceQuery = `
			SELECT ST_AsText(seg.geom), seg.length_m, seg.h3_r8::text
			FROM (
				SELECT (ST_Dump(ST_Force2D(ST_SetSRID(ST_GeomFromGeoJSON($1), 4326)))).geom AS way_line
			) dumped,
			LATERAL discretize_highway(dumped.way_line, 160.9344) seg
		`

		rows, err := tx.Query(ctx, sliceQuery, string(geomBytes))
		if err != nil {
			return stats, fmt.Errorf("discretize highway way %d: %w", hw.way.ID, err)
		}

		type slicedSegment struct {
			geomWKT string
			lengthM float64
			h3R8    string
		}
		var segments []slicedSegment
		for rows.Next() {
			var s slicedSegment
			if err := rows.Scan(&s.geomWKT, &s.lengthM, &s.h3R8); err != nil {
				rows.Close()
				return stats, fmt.Errorf("scan sliced segment: %w", err)
			}
			segments = append(segments, s)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return stats, fmt.Errorf("iterate sliced segments: %w", err)
		}

		startMP := hw.startMP
		if !hw.hasExplicitMP {
			startMP = corridorMP[hw.highwayRef]
		}
		currentMP := startMP
		for segIdx, seg := range segments {
			segStartMP := math.Round(currentMP*100) / 100
			deltaMiles := seg.lengthM / 1609.344
			currentMP += deltaMiles
			segEndMP := math.Round(currentMP*100) / 100

			nativeID := fmt.Sprintf("nps-%s-%d-%04d", hw.highwayRef, hw.way.ID, segIdx)
			osmWayIDs := []int64{hw.way.ID}

			const insertSegmentQuery = `
				INSERT INTO road_segments (
					source_id,
					native_id,
					geom,
					state,
					highway_ref,
					milepost_start,
					milepost_end,
					length_m,
					h3_r8,
					osm_way_ids
				) VALUES (
					$1, $2, ST_GeomFromText($3, 4326), $4, $5, $6, $7, $8, $9::h3index, $10
				)
				ON CONFLICT (source_id, native_id) DO UPDATE SET
					geom = EXCLUDED.geom,
					state = EXCLUDED.state,
					highway_ref = EXCLUDED.highway_ref,
					milepost_start = EXCLUDED.milepost_start,
					milepost_end = EXCLUDED.milepost_end,
					length_m = EXCLUDED.length_m,
					h3_r8 = EXCLUDED.h3_r8,
					osm_way_ids = EXCLUDED.osm_way_ids
			`

			_, err = tx.Exec(ctx, insertSegmentQuery,
				artifact.SourceID,
				nativeID,
				seg.geomWKT,
				hw.state,
				hw.highwayRef,
				segStartMP,
				segEndMP,
				seg.lengthM,
				seg.h3R8,
				osmWayIDs,
			)
			if err != nil {
				return stats, fmt.Errorf("insert road segment %s: %w", nativeID, err)
			}

			stats.SegmentsCreated++
			stats.TotalLengthM += seg.lengthM
		}

		corridorMP[hw.highwayRef] = currentMP
		stats.WaysParsed++
	}

	if err := tx.Commit(ctx); err != nil {
		return stats, fmt.Errorf("commit nps highway import: %w", err)
	}

	return stats, nil
}

func validateWGS84(lon, lat float64) error {
	if math.IsNaN(lon) || math.IsNaN(lat) || math.IsInf(lon, 0) || math.IsInf(lat, 0) {
		return errors.New("non-finite road coordinate")
	}
	if lon < -180 || lon > 180 || lat < -90 || lat > 90 {
		return errors.New("road coordinate out of WGS 84 bounds")
	}
	return nil
}

// inYellowstoneCorridorBbox checks if a coordinate falls within the Greater Yellowstone
// and Grand Teton highway pilot study area.
// Lat: [43.0, 46.0], Lon: [-112.5, -109.5].
func inYellowstoneCorridorBbox(lon, lat float64) bool {
	return lat >= 43.0 && lat <= 46.0 && lon >= -112.5 && lon <= -109.5
}

// matchHighwayRef identifies US-191 and US-89 corridors using exact tokens and regex boundaries.
func matchHighwayRef(tags map[string]string) (string, bool) {
	hw := tags["highway"]
	if hw == "" || hw == "footway" || hw == "path" || hw == "cycleway" || hw == "service" || hw == "track" || hw == "steps" {
		return "", false
	}

	ref := strings.TrimSpace(tags["ref"])
	name := strings.TrimSpace(tags["name"])

	has191 := isMatch191(ref, name)
	has89 := isMatch89(ref, name)

	if has191 && has89 {
		return "US-191; US-89", true
	}
	if has191 {
		return "US-191", true
	}
	if has89 {
		return "US-89", true
	}
	return "", false
}

func isMatch191(ref, name string) bool {
	if ref != "" {
		for _, part := range strings.Split(ref, ";") {
			p := strings.TrimSpace(strings.ToUpper(part))
			if p == "191" || p == "US 191" || p == "US-191" || p == "U.S. 191" || p == "U.S.-191" {
				return true
			}
		}
		if re191Ref.MatchString(ref) {
			return true
		}
	}
	if name != "" && re191Name.MatchString(name) {
		return true
	}
	return false
}

func isMatch89(ref, name string) bool {
	if ref != "" {
		for _, part := range strings.Split(ref, ";") {
			p := strings.TrimSpace(strings.ToUpper(part))
			if p == "89" || p == "US 89" || p == "US-89" || p == "U.S. 89" || p == "U.S.-89" {
				return true
			}
		}
		if re89Ref.MatchString(ref) {
			return true
		}
	}
	if name != "" && re89Name.MatchString(name) {
		return true
	}
	return false
}

// extractNodeMilepost extracts milestone values from node tags (including ref for milestone markers).
func extractNodeMilepost(tags map[string]string) (float64, bool) {
	for _, key := range []string{"milepost", "distance", "pk"} {
		if val, exists := tags[key]; exists {
			if mp, ok := parseMilepostNum(val); ok {
				return mp, true
			}
		}
	}
	if tags["highway"] == "milestone" {
		if val, exists := tags["ref"]; exists {
			if mp, ok := parseMilepostNum(val); ok {
				return mp, true
			}
		}
	}
	return 0, false
}

// extractWayMilepost extracts explicit starting milepost tags from highway ways.
// Crucially, it does NOT treat route 'ref' (e.g. ref="191" or ref="89") as a milepost.
func extractWayMilepost(tags map[string]string) (float64, bool) {
	for _, key := range []string{"milepost_start", "start_mp", "milepost"} {
		if val, exists := tags[key]; exists {
			if mp, ok := parseMilepostNum(val); ok {
				return mp, true
			}
		}
	}
	return 0, false
}

func parseMilepostNum(val string) (float64, bool) {
	val = strings.TrimPrefix(val, "MP")
	val = strings.TrimPrefix(val, "mp")
	val = strings.TrimSpace(val)
	mp, err := strconv.ParseFloat(val, 64)
	if err == nil && mp >= 0 && !math.IsNaN(mp) && !math.IsInf(mp, 0) {
		return mp, true
	}
	return 0, false
}

func determineState(tags map[string]string, lat, lon float64) string {
	if s := strings.ToUpper(strings.TrimSpace(tags["state"])); s == "WY" || s == "MT" || s == "ID" {
		return s
	}
	if s := strings.ToUpper(strings.TrimSpace(tags["addr:state"])); s == "WY" || s == "MT" || s == "ID" {
		return s
	}
	if lat >= 45.0 {
		return "MT"
	}
	if lon < -111.05 && lat < 44.5 {
		return "ID"
	}
	return "WY"
}
