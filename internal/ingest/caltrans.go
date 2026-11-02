package ingest

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"corridor/internal/db"
	"github.com/google/uuid"
)

// Caltrans FeatureServer service endpoints.
const (
	CaltransAADTFeatureServerURL      = "https://caltrans-gis.dot.ca.gov/arcgis/rest/services/CHhighway/Traffic_AADT/FeatureServer/0"
	CaltransPostmilesFeatureServerURL = "https://caltrans-gis.dot.ca.gov/arcgis/rest/services/CHhighway/SHN_Postmiles/FeatureServer/0"
)

// CaltransAADTArtifact returns the provenance artifact metadata for Caltrans AADT traffic counts.
func CaltransAADTArtifact() Artifact {
	return Artifact{
		SourceID:    "d8833219-913c-4435-8f2a-9a71bda57f76",
		Name:        "Caltrans · Annual Average Daily Traffic (AADT)",
		URL:         CaltransAADTFeatureServerURL,
		License:     "CC-BY · State of California / Caltrans",
		SHA256:      "",
		ObjectKey:   "",
		RetrievedAt: time.Date(2026, 9, 21, 21, 4, 41, 0, time.UTC),
	}
}

// CaltransPostmilesArtifact returns the provenance artifact metadata for Caltrans State Highway Network (SHN) Postmiles LRS.
func CaltransPostmilesArtifact() Artifact {
	return Artifact{
		SourceID:    "c7348e3a-921c-43df-9642-a162d0f04c62",
		Name:        "Caltrans · State Highway Network (SHN) Postmiles LRS",
		URL:         CaltransPostmilesFeatureServerURL,
		License:     "Public Use · State of California / Caltrans",
		SHA256:      "",
		ObjectKey:   "",
		RetrievedAt: time.Date(2026, 9, 21, 21, 4, 41, 0, time.UTC),
	}
}

// CaltransAADTRecord represents an Annual Average Daily Traffic (AADT) count station record from Caltrans.
type CaltransAADTRecord struct {
	District       string  `json:"district"`
	Route          string  `json:"route"`
	County         string  `json:"county"`
	Postmile       float64 `json:"postmile"`
	PostmilePrefix string  `json:"postmile_prefix,omitempty"`
	PostmileSuffix string  `json:"postmile_suffix,omitempty"`
	Description    string  `json:"description,omitempty"`
	BackAADT       float64 `json:"back_aadt"`
	AheadAADT      float64 `json:"ahead_aadt"`
	BackPeakHour   float64 `json:"back_peak_hour"`
	AheadPeakHour  float64 `json:"ahead_peak_hour"`
	Longitude      float64 `json:"longitude,omitempty"`
	Latitude       float64 `json:"latitude,omitempty"`
}

// AverageBidirectionalAADT calculates the average bidirectional AADT for this count station.
// If both BackAADT and AheadAADT are positive, it returns their arithmetic mean.
// If only AheadAADT is positive (e.g., at route start or forward branch), it returns AheadAADT.
// If only BackAADT is positive (e.g., at route terminus or backward branch), it returns BackAADT.
// Otherwise, it returns 0.
func (r CaltransAADTRecord) AverageBidirectionalAADT() float64 {
	if r.BackAADT > 0 && r.AheadAADT > 0 {
		return (r.BackAADT + r.AheadAADT) / 2.0
	}
	if r.AheadAADT > 0 {
		return r.AheadAADT
	}
	if r.BackAADT > 0 {
		return r.BackAADT
	}
	return 0
}

// AverageAADT is an alias for AverageBidirectionalAADT.
func (r CaltransAADTRecord) AverageAADT() float64 {
	return r.AverageBidirectionalAADT()
}

// AveragePeakHour calculates the average peak hour traffic volume for this count station.
func (r CaltransAADTRecord) AveragePeakHour() float64 {
	if r.BackPeakHour > 0 && r.AheadPeakHour > 0 {
		return (r.BackPeakHour + r.AheadPeakHour) / 2.0
	}
	if r.AheadPeakHour > 0 {
		return r.AheadPeakHour
	}
	if r.BackPeakHour > 0 {
		return r.BackPeakHour
	}
	return 0
}

// DistrictNum returns the integer district number, or 0 if unparseable.
func (r CaltransAADTRecord) DistrictNum() int {
	d, _ := strconv.Atoi(strings.TrimSpace(r.District))
	return d
}

// CaltransPostmileRecord represents a linear referencing postmile feature from Caltrans SHN_Postmiles.
type CaltransPostmileRecord struct {
	ObjectID    int64        `json:"object_id"`
	RouteID     string       `json:"route_id"`
	County      string       `json:"county"`
	RouteNum    string       `json:"route_num"`
	RouteSuffix string       `json:"route_suffix,omitempty"`
	PMPrefix    string       `json:"pm_prefix,omitempty"`
	PMSuffix    string       `json:"pm_suffix,omitempty"`
	Alignment   string       `json:"alignment,omitempty"`
	LineID      string       `json:"line_id,omitempty"`
	LineOrder   int          `json:"line_order,omitempty"`
	ShapeLength float64      `json:"shape_length,omitempty"`
	Coordinates [][2]float64 `json:"coordinates,omitempty"`
}

// PostmileReference represents a parsed Caltrans postmile identifier (e.g. "10-MPA-120-R43.09").
type PostmileReference struct {
	District string  `json:"district,omitempty"`
	County   string  `json:"county,omitempty"`
	Route    string  `json:"route"`
	Prefix   string  `json:"prefix,omitempty"`
	Postmile float64 `json:"postmile"`
	Suffix   string  `json:"suffix,omitempty"`
}

var (
	// Matches "10-MPA-120-R43.09" or "MPA-120-43.09" or "10-MPA-120-43.09A"
	reCaltransPostmileStandard = regexp.MustCompile(`^(?:(\d{1,2})-)?([A-Za-z]{2,3})-([A-Za-z0-9]+)-([A-Za-z])?([0-9]+(?:\.[0-9]+)?)([A-Za-z])?$`)
	// Matches "CA-120 43.09" or "SR-120 PM 43.09" or "120-43.09"
	reCaltransPostmileSimple = regexp.MustCompile(`^(?:(?:CA|SR|State Route|Highway)[- ]*)?([A-Za-z0-9]+)[- ]+(?:PM[- ]*)?([A-Za-z])?([0-9]+(?:\.[0-9]+)?)([A-Za-z])?$`)
)

// FormatPostmile formats a Caltrans Postmile reference string.
// Example: FormatPostmile("10", "MPA", "120", "R", 43.09, "") -> "10-MPA-120-R43.09".
func FormatPostmile(district, county, route, prefix string, pm float64, suffix string) string {
	var sb strings.Builder
	if district != "" {
		sb.WriteString(strings.TrimSpace(district))
		sb.WriteString("-")
	}
	if county != "" {
		sb.WriteString(strings.ToUpper(strings.TrimSpace(county)))
		sb.WriteString("-")
	}
	sb.WriteString(strings.ToUpper(strings.TrimSpace(route)))
	sb.WriteString("-")
	if prefix != "" {
		sb.WriteString(strings.ToUpper(strings.TrimSpace(prefix)))
	}
	sb.WriteString(strconv.FormatFloat(pm, 'f', -1, 64))
	if suffix != "" {
		sb.WriteString(strings.ToUpper(strings.TrimSpace(suffix)))
	}
	return sb.String()
}

// ParsePostmileReference parses a Caltrans postmile designation string into its structured components.
func ParsePostmileReference(s string) (PostmileReference, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return PostmileReference{}, errors.New("empty postmile string")
	}

	if m := reCaltransPostmileStandard.FindStringSubmatch(s); m != nil {
		pm, err := strconv.ParseFloat(m[5], 64)
		if err != nil {
			return PostmileReference{}, fmt.Errorf("parse postmile value %q: %w", m[5], err)
		}
		return PostmileReference{
			District: m[1],
			County:   strings.ToUpper(m[2]),
			Route:    strings.ToUpper(m[3]),
			Prefix:   strings.ToUpper(m[4]),
			Postmile: pm,
			Suffix:   strings.ToUpper(m[6]),
		}, nil
	}

	if m := reCaltransPostmileSimple.FindStringSubmatch(s); m != nil {
		pm, err := strconv.ParseFloat(m[3], 64)
		if err != nil {
			return PostmileReference{}, fmt.Errorf("parse postmile value %q: %w", m[3], err)
		}
		return PostmileReference{
			Route:    strings.ToUpper(m[1]),
			Prefix:   strings.ToUpper(m[2]),
			Postmile: pm,
			Suffix:   strings.ToUpper(m[4]),
		}, nil
	}

	return PostmileReference{}, fmt.Errorf("unrecognized Caltrans postmile format: %q", s)
}

// ParseCaltrans parses Caltrans traffic and roadway data from an io.Reader (CSV or GeoJSON).
func ParseCaltrans(r io.Reader) ([]CaltransAADTRecord, error) {
	return ParseCaltransAADT(r)
}

// ParseCaltransAADT parses Caltrans Annual Average Daily Traffic records from an io.Reader.
// It auto-detects GeoJSON vs CSV format based on the stream's leading non-whitespace byte.
func ParseCaltransAADT(r io.Reader) ([]CaltransAADTRecord, error) {
	if r == nil {
		return nil, errors.New("nil reader")
	}

	br := bufio.NewReader(io.LimitReader(r, 64<<20))
	firstByte, err := peekFirstNonWhitespace(br)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("empty input")
		}
		return nil, fmt.Errorf("read header: %w", err)
	}

	if firstByte == '{' || firstByte == '[' {
		return parseCaltransAADTGeoJSON(br)
	}
	return parseCaltransAADTCSV(br)
}

func peekFirstNonWhitespace(br *bufio.Reader) (byte, error) {
	for {
		b, err := br.ReadByte()
		if err != nil {
			return 0, err
		}
		if b != ' ' && b != '\t' && b != '\r' && b != '\n' {
			if err := br.UnreadByte(); err != nil {
				return 0, err
			}
			return b, nil
		}
	}
}

type geoJSONDoc struct {
	Type     string           `json:"type"`
	Features []geoJSONFeature `json:"features"`
}

type geoJSONFeature struct {
	Type       string                 `json:"type"`
	ID         any                    `json:"id"`
	Geometry   *geoJSONGeom           `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type geoJSONGeom struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

func parseCaltransAADTGeoJSON(r io.Reader) ([]CaltransAADTRecord, error) {
	dec := json.NewDecoder(r)
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode caltrans geojson: %w", err)
	}

	var doc geoJSONDoc
	if err := json.Unmarshal(raw, &doc); err == nil && (doc.Type == "FeatureCollection" || len(doc.Features) > 0) {
		return convertGeoJSONFeaturesToAADT(doc.Features)
	}

	// Try single Feature
	var singleFeature geoJSONFeature
	if err := json.Unmarshal(raw, &singleFeature); err == nil && singleFeature.Type == "Feature" {
		return convertGeoJSONFeaturesToAADT([]geoJSONFeature{singleFeature})
	}

	// Try slice of features
	var featureSlice []geoJSONFeature
	if err := json.Unmarshal(raw, &featureSlice); err == nil && len(featureSlice) > 0 {
		return convertGeoJSONFeaturesToAADT(featureSlice)
	}

	return nil, errors.New("invalid caltrans GeoJSON: missing FeatureCollection or Feature")
}

func convertGeoJSONFeaturesToAADT(features []geoJSONFeature) ([]CaltransAADTRecord, error) {
	records := make([]CaltransAADTRecord, 0, len(features))
	for i, f := range features {
		rec, err := parseAADTFeature(f)
		if err != nil {
			return nil, fmt.Errorf("feature %d: %w", i, err)
		}
		records = append(records, rec)
	}
	return records, nil
}

func parseAADTFeature(f geoJSONFeature) (CaltransAADTRecord, error) {
	props := f.Properties
	if props == nil {
		return CaltransAADTRecord{}, errors.New("feature missing properties")
	}

	var rec CaltransAADTRecord

	// Route
	if val, ok := getCaltransProperty(props, "route", "rte", "routenum", "highway"); ok {
		rec.Route = parseCaltransString(val)
	}
	if rec.Route == "" {
		return CaltransAADTRecord{}, errors.New("missing route")
	}

	// District
	if val, ok := getCaltransProperty(props, "district", "dist", "districtnum"); ok {
		rec.District = parseCaltransString(val)
	}

	// County
	if val, ok := getCaltransProperty(props, "county", "cnty", "countycode", "cty"); ok {
		rec.County = strings.ToUpper(parseCaltransString(val))
	}

	// Postmile
	if val, ok := getCaltransProperty(props, "postmile", "pm", "milepost", "mp"); ok {
		pm, err := parseCaltransFloat(val)
		if err != nil {
			return CaltransAADTRecord{}, fmt.Errorf("invalid postmile: %w", err)
		}
		rec.Postmile = pm
	} else {
		return CaltransAADTRecord{}, errors.New("missing postmile")
	}

	// Postmile Prefix & Suffix
	if val, ok := getCaltransProperty(props, "pmpfx", "pm_pfx", "prefix", "pmprefix"); ok {
		rec.PostmilePrefix = parseCaltransString(val)
	}
	if val, ok := getCaltransProperty(props, "pmsfx", "pm_sfx", "suffix", "pmsuffix"); ok {
		rec.PostmileSuffix = parseCaltransString(val)
	}

	// Description
	if val, ok := getCaltransProperty(props, "description", "desc", "location"); ok {
		rec.Description = parseCaltransString(val)
	}

	// Back AADT
	if val, ok := getCaltransProperty(props, "backaadt", "back_aadt", "backadt", "back_adt"); ok {
		v, err := parseCaltransFloat(val)
		if err != nil {
			return CaltransAADTRecord{}, fmt.Errorf("invalid BackAADT: %w", err)
		}
		rec.BackAADT = v
	}

	// Ahead AADT
	if val, ok := getCaltransProperty(props, "aheadaadt", "ahead_aadt", "aheadadt", "ahead_adt"); ok {
		v, err := parseCaltransFloat(val)
		if err != nil {
			return CaltransAADTRecord{}, fmt.Errorf("invalid AheadAADT: %w", err)
		}
		rec.AheadAADT = v
	}

	// Back Peak Hour
	if val, ok := getCaltransProperty(props, "backpeakhour", "back_peak_hour", "backpeak", "bpeakhour"); ok {
		v, err := parseCaltransFloat(val)
		if err != nil {
			return CaltransAADTRecord{}, fmt.Errorf("invalid BackPeakHour: %w", err)
		}
		rec.BackPeakHour = v
	}

	// Ahead Peak Hour
	if val, ok := getCaltransProperty(props, "aheadpeakhour", "ahead_peak_hour", "aheadpeak", "apeakhour"); ok {
		v, err := parseCaltransFloat(val)
		if err != nil {
			return CaltransAADTRecord{}, fmt.Errorf("invalid AheadPeakHour: %w", err)
		}
		rec.AheadPeakHour = v
	}

	// Geometry coordinates
	if f.Geometry != nil && strings.EqualFold(f.Geometry.Type, "Point") && len(f.Geometry.Coordinates) > 0 {
		var pt []float64
		if err := json.Unmarshal(f.Geometry.Coordinates, &pt); err == nil && len(pt) >= 2 {
			lon, lat := pt[0], pt[1]
			if lon >= -180 && lon <= 180 && lat >= -90 && lat <= 90 {
				rec.Longitude = lon
				rec.Latitude = lat
			} else {
				return CaltransAADTRecord{}, fmt.Errorf("invalid point coordinates: [%f, %f]", lon, lat)
			}
		}
	}

	// Property coordinate fallback
	if rec.Longitude == 0 && rec.Latitude == 0 {
		if xVal, ok := getCaltransProperty(props, "x", "lon", "long", "longitude"); ok {
			if yVal, ok := getCaltransProperty(props, "y", "lat", "latitude"); ok {
				lon, err1 := parseCaltransFloat(xVal)
				lat, err2 := parseCaltransFloat(yVal)
				if err1 == nil && err2 == nil && lon >= -180 && lon <= 180 && lat >= -90 && lat <= 90 {
					rec.Longitude = lon
					rec.Latitude = lat
				}
			}
		}
	}

	return rec, nil
}

func parseCaltransAADTCSV(r io.Reader) ([]CaltransAADTRecord, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}

	idxDistrict := -1
	idxRoute := -1
	idxCounty := -1
	idxPostmile := -1
	idxPMPfx := -1
	idxPMSfx := -1
	idxDesc := -1
	idxBackAADT := -1
	idxAheadAADT := -1
	idxBackPeak := -1
	idxAheadPeak := -1
	idxLon := -1
	idxLat := -1

	for i, col := range header {
		norm := normalizeCaltransKey(col)
		switch norm {
		case "district", "dist", "districtnum":
			idxDistrict = i
		case "route", "rte", "routenum", "highway":
			idxRoute = i
		case "county", "cnty", "countycode", "cty":
			idxCounty = i
		case "postmile", "pm", "milepost", "mp":
			idxPostmile = i
		case "pmpfx", "pmprefix", "postmileprefix", "prefix":
			idxPMPfx = i
		case "pmsfx", "pmsuffix", "postmilesuffix", "suffix":
			idxPMSfx = i
		case "description", "desc", "location":
			idxDesc = i
		case "backaadt", "backadt", "backaadtvolume":
			idxBackAADT = i
		case "aheadaadt", "aheadadt", "aheadaadtvolume":
			idxAheadAADT = i
		case "backpeakhour", "backpeak", "bpeakhour":
			idxBackPeak = i
		case "aheadpeakhour", "aheadpeak", "apeakhour":
			idxAheadPeak = i
		case "longitude", "lon", "long", "x":
			idxLon = i
		case "latitude", "lat", "y":
			idxLat = i
		}
	}

	if idxRoute == -1 || idxPostmile == -1 {
		return nil, errors.New("missing required columns in Caltrans AADT CSV (Route and Postmile required)")
	}

	var records []CaltransAADTRecord
	for lineNum := 2; ; lineNum++ {
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read CSV row %d: %w", lineNum, err)
		}
		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}

		var rec CaltransAADTRecord

		// Route
		if idxRoute < len(row) {
			rec.Route = strings.TrimSpace(row[idxRoute])
		}
		if rec.Route == "" {
			return nil, fmt.Errorf("row %d: missing route", lineNum)
		}

		// District
		if idxDistrict != -1 && idxDistrict < len(row) {
			rec.District = strings.TrimSpace(row[idxDistrict])
		}

		// County
		if idxCounty != -1 && idxCounty < len(row) {
			rec.County = strings.ToUpper(strings.TrimSpace(row[idxCounty]))
		}

		// Postmile
		if idxPostmile < len(row) {
			pmStr := strings.TrimSpace(row[idxPostmile])
			if pmStr == "" {
				return nil, fmt.Errorf("row %d: missing postmile", lineNum)
			}
			pm, err := strconv.ParseFloat(pmStr, 64)
			if err != nil {
				return nil, fmt.Errorf("row %d: invalid postmile %q: %w", lineNum, pmStr, err)
			}
			rec.Postmile = pm
		}

		// Postmile Prefix & Suffix
		if idxPMPfx != -1 && idxPMPfx < len(row) {
			rec.PostmilePrefix = strings.TrimSpace(row[idxPMPfx])
		}
		if idxPMSfx != -1 && idxPMSfx < len(row) {
			rec.PostmileSuffix = strings.TrimSpace(row[idxPMSfx])
		}

		// Description
		if idxDesc != -1 && idxDesc < len(row) {
			rec.Description = strings.TrimSpace(row[idxDesc])
		}

		// Back AADT
		if idxBackAADT != -1 && idxBackAADT < len(row) {
			s := strings.ReplaceAll(strings.TrimSpace(row[idxBackAADT]), ",", "")
			if s != "" && s != "-" && !strings.EqualFold(s, "N/A") {
				v, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return nil, fmt.Errorf("row %d: invalid BackAADT %q: %w", lineNum, s, err)
				}
				rec.BackAADT = v
			}
		}

		// Ahead AADT
		if idxAheadAADT != -1 && idxAheadAADT < len(row) {
			s := strings.ReplaceAll(strings.TrimSpace(row[idxAheadAADT]), ",", "")
			if s != "" && s != "-" && !strings.EqualFold(s, "N/A") {
				v, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return nil, fmt.Errorf("row %d: invalid AheadAADT %q: %w", lineNum, s, err)
				}
				rec.AheadAADT = v
			}
		}

		// Back Peak Hour
		if idxBackPeak != -1 && idxBackPeak < len(row) {
			s := strings.ReplaceAll(strings.TrimSpace(row[idxBackPeak]), ",", "")
			if s != "" && s != "-" && !strings.EqualFold(s, "N/A") {
				v, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return nil, fmt.Errorf("row %d: invalid BackPeakHour %q: %w", lineNum, s, err)
				}
				rec.BackPeakHour = v
			}
		}

		// Ahead Peak Hour
		if idxAheadPeak != -1 && idxAheadPeak < len(row) {
			s := strings.ReplaceAll(strings.TrimSpace(row[idxAheadPeak]), ",", "")
			if s != "" && s != "-" && !strings.EqualFold(s, "N/A") {
				v, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return nil, fmt.Errorf("row %d: invalid AheadPeakHour %q: %w", lineNum, s, err)
				}
				rec.AheadPeakHour = v
			}
		}

		// Longitude
		if idxLon != -1 && idxLon < len(row) {
			s := strings.TrimSpace(row[idxLon])
			if s != "" {
				lon, err := strconv.ParseFloat(s, 64)
				if err == nil && lon >= -180 && lon <= 180 {
					rec.Longitude = lon
				}
			}
		}

		// Latitude
		if idxLat != -1 && idxLat < len(row) {
			s := strings.TrimSpace(row[idxLat])
			if s != "" {
				lat, err := strconv.ParseFloat(s, 64)
				if err == nil && lat >= -90 && lat <= 90 {
					rec.Latitude = lat
				}
			}
		}

		records = append(records, rec)
	}

	return records, nil
}

// ParseCaltransPostmiles parses linear postmile features from the Caltrans SHN_Postmiles FeatureServer GeoJSON export.
func ParseCaltransPostmiles(r io.Reader) ([]CaltransPostmileRecord, error) {
	if r == nil {
		return nil, errors.New("nil reader")
	}

	var doc struct {
		Type     string `json:"type"`
		Features []struct {
			Type     string                 `json:"type"`
			ID       int64                  `json:"id"`
			Geometry geoJSONGeom            `json:"geometry"`
			Props    map[string]interface{} `json:"properties"`
		} `json:"features"`
	}

	if err := json.NewDecoder(io.LimitReader(r, 64<<20)).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode postmiles geojson: %w", err)
	}

	records := make([]CaltransPostmileRecord, 0, len(doc.Features))
	for i, f := range doc.Features {
		props := f.Props
		rec := CaltransPostmileRecord{
			ObjectID: f.ID,
		}
		if oidVal, ok := getCaltransProperty(props, "objectid"); ok {
			if v, err := parseCaltransFloat(oidVal); err == nil {
				rec.ObjectID = int64(v)
			}
		}
		if v, ok := getCaltransProperty(props, "routeid"); ok {
			rec.RouteID = parseCaltransString(v)
		}
		if v, ok := getCaltransProperty(props, "county"); ok {
			rec.County = strings.ToUpper(parseCaltransString(v))
		}
		if v, ok := getCaltransProperty(props, "routenum", "route"); ok {
			rec.RouteNum = parseCaltransString(v)
		}
		if v, ok := getCaltransProperty(props, "routesuffix"); ok {
			rec.RouteSuffix = parseCaltransString(v)
		}
		if v, ok := getCaltransProperty(props, "pmprefix"); ok {
			rec.PMPrefix = parseCaltransString(v)
		}
		if v, ok := getCaltransProperty(props, "pmsuffix"); ok {
			rec.PMSuffix = parseCaltransString(v)
		}
		if v, ok := getCaltransProperty(props, "alignment"); ok {
			rec.Alignment = parseCaltransString(v)
		}
		if v, ok := getCaltransProperty(props, "lineid"); ok {
			rec.LineID = parseCaltransString(v)
		}
		if v, ok := getCaltransProperty(props, "lineorder"); ok {
			if num, err := parseCaltransFloat(v); err == nil {
				rec.LineOrder = int(num)
			}
		}
		if v, ok := getCaltransProperty(props, "shape__length", "shapelength"); ok {
			if num, err := parseCaltransFloat(v); err == nil {
				rec.ShapeLength = num
			}
		}

		// Geometry parsing (LineString or MultiLineString)
		if strings.EqualFold(f.Geometry.Type, "LineString") {
			var coords [][2]float64
			if err := json.Unmarshal(f.Geometry.Coordinates, &coords); err == nil {
				rec.Coordinates = coords
			}
		} else if strings.EqualFold(f.Geometry.Type, "MultiLineString") {
			var multi [][][2]float64
			if err := json.Unmarshal(f.Geometry.Coordinates, &multi); err == nil && len(multi) > 0 {
				rec.Coordinates = multi[0]
			}
		}

		if rec.RouteNum == "" && rec.RouteID == "" {
			return nil, fmt.Errorf("feature %d: missing route identifier", i)
		}

		records = append(records, rec)
	}

	return records, nil
}

// MapAADTToSegments maps Caltrans AADT count records to road segments, calculating the average
// bidirectional AADT for each segment based on route matching and spatial proximity.
func MapAADTToSegments(records []CaltransAADTRecord, segments []db.SegmentSummary) map[uuid.UUID]float64 {
	result := make(map[uuid.UUID]float64, len(segments))
	if len(records) == 0 || len(segments) == 0 {
		return result
	}

	// Group records by normalized route
	recordsByRoute := make(map[string][]CaltransAADTRecord)
	for _, r := range records {
		norm := normalizeCaltransRoute(r.Route)
		recordsByRoute[norm] = append(recordsByRoute[norm], r)
	}

	for _, seg := range segments {
		var segRoute string
		if seg.HighwayRef != nil {
			segRoute = normalizeCaltransRoute(*seg.HighwayRef)
		}

		// Select candidate records for this segment
		candidates := recordsByRoute[segRoute]
		if len(candidates) == 0 {
			if segRoute == "" {
				// No route specified on segment; consider all records
				candidates = records
			} else {
				// HighwayRef specified but no matching records found
				continue
			}
		}

		// Determine if candidate records have geographic coordinates
		hasCoords := false
		for _, c := range candidates {
			if c.Longitude != 0 || c.Latitude != 0 {
				hasCoords = true
				break
			}
		}

		if hasCoords && len(seg.Coordinates) > 0 {
			// Find nearest candidate count station along the route
			minDist := math.MaxFloat64
			var bestRec CaltransAADTRecord
			found := false

			for _, c := range candidates {
				if c.Longitude == 0 && c.Latitude == 0 {
					continue
				}
				d := minDistanceToSegment(c.Longitude, c.Latitude, seg.Coordinates)
				if d < minDist {
					minDist = d
					bestRec = c
					found = true
				}
			}

			if found {
				result[seg.ID] = bestRec.AverageBidirectionalAADT()
				continue
			}
		}

		// Fallback: calculate arithmetic mean of bidirectional AADT across matching route candidates
		var sum float64
		var count int
		for _, c := range candidates {
			aadt := c.AverageBidirectionalAADT()
			if aadt > 0 {
				sum += aadt
				count++
			}
		}

		if count > 0 {
			result[seg.ID] = sum / float64(count)
		} else {
			result[seg.ID] = 0
		}
	}

	return result
}

func minDistanceToSegment(recLon, recLat float64, segCoords [][]float64) float64 {
	minDist := math.MaxFloat64
	for _, pt := range segCoords {
		if len(pt) >= 2 {
			d := haversineDistanceMeters(recLon, recLat, pt[0], pt[1])
			if d < minDist {
				minDist = d
			}
		}
	}
	return minDist
}

func haversineDistanceMeters(lon1, lat1, lon2, lat2 float64) float64 {
	const earthRadiusM = 6371000.0
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	lat1Rad := lat1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

func normalizeCaltransRoute(route string) string {
	r := strings.ToUpper(strings.TrimSpace(route))
	r = strings.TrimPrefix(r, "STATE ROUTE ")
	r = strings.TrimPrefix(r, "HIGHWAY ")
	r = strings.TrimPrefix(r, "HWY ")
	r = strings.TrimPrefix(r, "CA-")
	r = strings.TrimPrefix(r, "CA ")
	r = strings.TrimPrefix(r, "SR-")
	r = strings.TrimPrefix(r, "SR ")
	r = strings.TrimPrefix(r, "US-")
	r = strings.TrimPrefix(r, "US ")
	r = strings.TrimPrefix(r, "INTERSTATE ")
	r = strings.TrimPrefix(r, "I-")
	r = strings.TrimPrefix(r, "I ")
	return strings.TrimSpace(r)
}

func normalizeCaltransKey(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

func getCaltransProperty(props map[string]interface{}, aliases ...string) (any, bool) {
	for _, alias := range aliases {
		normAlias := normalizeCaltransKey(alias)
		for k, v := range props {
			if normalizeCaltransKey(k) == normAlias {
				return v, true
			}
		}
	}
	return nil, false
}

func parseCaltransString(val any) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		if v == math.Trunc(v) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func parseCaltransFloat(val any) (float64, error) {
	if val == nil {
		return 0, nil
	}
	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		s := strings.ReplaceAll(strings.TrimSpace(v), ",", "")
		if s == "" || s == "-" || strings.EqualFold(s, "N/A") || strings.EqualFold(s, "NA") {
			return 0, nil
		}
		num, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}
		return num, nil
	default:
		s := fmt.Sprintf("%v", val)
		return strconv.ParseFloat(s, 64)
	}
}
