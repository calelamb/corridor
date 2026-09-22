package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Road struct {
	ID       string
	Ref      string
	Geometry json.RawMessage
}
type roadPoint struct {
	Lon *float64 `json:"lon"`
	Lat *float64 `json:"lat"`
}
type roadWay struct {
	Type     string            `json:"type"`
	ID       int64             `json:"id"`
	Tags     map[string]string `json:"tags"`
	Geometry []roadPoint       `json:"geometry"`
}

func RoadArtifact() Artifact {
	const hash = "0dafb47436631d57703fd21ce74eca160a91d991852dfea98bb28a1f4d70e659"
	return Artifact{SourceID: "48d59a58-9b6b-4f1a-a1fa-ce0cdedc1f81", Name: "OpenStreetMap · Pequop I-80", URL: "https://www.openstreetmap.org/copyright", License: "ODbL 1.0 · © OpenStreetMap contributors", SHA256: hash, ObjectKey: "sha256/" + hash + "/pequop-roads.json", RetrievedAt: time.Date(2026, 9, 21, 22, 33, 34, 0, time.UTC)}
}
func ParseRoads(r io.Reader) ([]Road, error) {
	var document struct {
		Remark   string    `json:"remark"`
		Elements []roadWay `json:"elements"`
	}
	decoder := json.NewDecoder(io.LimitReader(r, 8<<20))
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decode road source: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, errors.New("trailing road data")
	}
	if document.Remark != "" || len(document.Elements) == 0 || len(document.Elements) > 5000 {
		return nil, errors.New("incomplete or excessive road source")
	}
	roads := make([]Road, 0, len(document.Elements))
	seen := map[int64]bool{}
	for _, way := range document.Elements {
		if seen[way.ID] {
			return nil, errors.New("duplicate road way")
		}
		road, err := parseWay(way)
		if err != nil {
			return nil, err
		}
		seen[way.ID] = true
		roads = append(roads, road)
	}
	return roads, nil
}
func parseWay(w roadWay) (Road, error) {
	if w.Type != "way" || w.ID <= 0 || w.Tags["highway"] != "motorway" || !strings.Contains(w.Tags["ref"], "I 80") || len(w.Tags["ref"]) > 100 || len(w.Geometry) < 2 || len(w.Geometry) > 50000 {
		return Road{}, errors.New("invalid I-80 road way")
	}
	points := make([][2]float64, 0, len(w.Geometry))
	for _, p := range w.Geometry {
		if p.Lon == nil || p.Lat == nil || math.IsNaN(*p.Lon) || math.IsNaN(*p.Lat) || math.Abs(*p.Lon) > 180 || math.Abs(*p.Lat) > 85 {
			return Road{}, errors.New("invalid road coordinate")
		}
		points = append(points, [2]float64{*p.Lon, *p.Lat})
	}
	geom, err := json.Marshal(struct {
		Type        string       `json:"type"`
		Coordinates [][2]float64 `json:"coordinates"`
	}{"LineString", points})
	if err != nil {
		return Road{}, err
	}
	return Road{strconv.FormatInt(w.ID, 10), w.Tags["ref"], geom}, nil
}
func ImportRoads(ctx context.Context, pool *pgxpool.Pool, roads []Road, a Artifact) error {
	if len(roads) == 0 {
		return errors.New("no roads to import")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(827367)"); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO sources(id,name,url,license,status,public_url) VALUES($1,$2,$3,$4,'approved',$3) ON CONFLICT DO NOTHING`, a.SourceID, a.Name, a.URL, a.License); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO source_artifacts(source_id,object_key,retrieved_at,sha256) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, a.SourceID, a.ObjectKey, a.RetrievedAt, a.SHA256); err != nil {
		return err
	}
	for _, r := range roads {
		if _, err = tx.Exec(ctx, `INSERT INTO analysis_roads(source_id,native_id,ref,geom) VALUES($1,$2,$3,ST_SetSRID(ST_GeomFromGeoJSON($4),4326)) ON CONFLICT DO NOTHING`, a.SourceID, r.ID, r.Ref, string(r.Geometry)); err != nil {
			return fmt.Errorf("insert road: %w", err)
		}
	}
	if _, err = tx.Exec(ctx, `SELECT rebuild_movement_features()`); err != nil {
		return fmt.Errorf("build movement support: %w", err)
	}
	return tx.Commit(ctx)
}
