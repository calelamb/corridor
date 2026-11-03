// Package ingest normalizes licensed source artifacts without publishing raw records.
package ingest

import (
	"bytes"
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

	"corridor/internal/geo"
)

// MDTCarcassArtifact returns provenance metadata for the Montana Department of
// Transportation statewide carcass collection cataloged as S32.
//
// SHA256 stays empty until CarcassData2016-2025.xlsx is acquired in this
// workspace. data/SHA256SUMS does not list that workbook, and the file is not
// in the repository, so a previously recorded digest cannot be re-verified.
// RetrievedAt is the retrieval timestamp already recorded in
// data/sweeps/2026-09-21/artifacts.json. This function does not download the file.
func MDTCarcassArtifact() Artifact {
	return Artifact{
		SourceID:    "9f43d083-c82b-a360-eb71-74815e633f04",
		Name:        "MDT · Statewide Carcass Collection 2016-2025",
		URL:         "https://www.mdt.mt.gov/publications/docs/datastats/crashdata/CarcassData2016-2025.xlsx",
		License:     "State Government Public Record · Montana Department of Transportation",
		SHA256:      "",
		ObjectKey:   "",
		RetrievedAt: time.Date(2026, 9, 21, 21, 8, 51, 0, time.UTC),
	}
}

// MDTCarcassRecord represents a parsed, normalized carcass salvage / pickup event
// recorded along the Montana State Highway Network by MDT maintenance personnel.
type MDTCarcassRecord struct {
	CollectionDate    time.Time `json:"collection_date"`
	Route             string    `json:"route"`
	Milepost          float64   `json:"milepost"`
	AnimalType        string    `json:"animal_type"`
	NormalizedSpecies string    `json:"normalized_species"`
	Gender            string    `json:"gender"`
	Quantity          int       `json:"quantity"`
	SMSX              float64   `json:"-"` // State-plane easting. JSON would reconstruct a precise point.
	SMSY              float64   `json:"-"` // State-plane northing. JSON would reconstruct a precise point.
	SMS_X_CORD        float64   `json:"-"`
	SMS_Y_CORD        float64   `json:"-"`
	Longitude         float64   `json:"longitude,omitempty"` // Public WGS84 longitude (H3-r6 centroid when sensitive).
	Latitude          float64   `json:"latitude,omitempty"`  // Public WGS84 latitude (H3-r6 centroid when sensitive).
	RawLongitude      float64   `json:"-"`
	RawLatitude       float64   `json:"-"`
	IsSensitive       bool      `json:"is_sensitive"`
	H3R6              string    `json:"h3_r6,omitempty"`
	H3Index           string    `json:"h3_index,omitempty"`
	QuarantineUntil   time.Time `json:"quarantine_until,omitempty"`
	InQuarantine      bool      `json:"in_quarantine"`
}

// mdtCanonicalRoute matches a Montana highway class and number after separators are removed.
var mdtCanonicalRoute = regexp.MustCompile(`^(I|US|MT|S|P|SR)-?(\d+)([A-Z])?$`)

// NormalizeMDTRoute canonicalizes MDT highway labels such as "us 191" and "I-90".
// Labels that are not a known highway class are returned trimmed and uppercased.
func NormalizeMDTRoute(raw string) string {
	clean := strings.ToUpper(strings.TrimSpace(raw))
	if clean == "" || clean == "UNKNOWN" || clean == "N/A" || clean == "NA" {
		return "Unknown"
	}
	clean = strings.ReplaceAll(clean, ".", "")
	clean = strings.ReplaceAll(clean, "_", " ")
	clean = strings.Join(strings.Fields(clean), " ")
	compact := strings.ReplaceAll(clean, " ", "")
	compact = strings.ReplaceAll(compact, "–", "-")
	compact = strings.ReplaceAll(compact, "—", "-")
	switch {
	case strings.HasPrefix(compact, "INTERSTATE"):
		compact = "I" + strings.TrimPrefix(compact, "INTERSTATE")
	case strings.HasPrefix(compact, "MONTANA"):
		compact = "MT" + strings.TrimPrefix(compact, "MONTANA")
	}
	match := mdtCanonicalRoute.FindStringSubmatch(compact)
	if match == nil {
		return clean
	}
	number := strings.TrimLeft(match[2], "0")
	if number == "" {
		number = "0"
	}
	return match[1] + "-" + number + match[3]
}

// InQuarantineAt reports whether a sensitive record is still inside the 30-day public delay.
func (r MDTCarcassRecord) InQuarantineAt(refTime time.Time) bool {
	if !r.IsSensitive || r.QuarantineUntil.IsZero() {
		return false
	}
	return refTime.Before(r.QuarantineUntil)
}

// PublicRecord returns the privacy-compliant record for public feeds.
// Sensitive records inside the 30-day delay are withheld.
// Released sensitive records keep the H3 resolution 6 centroid and drop precise coordinates.
func (r MDTCarcassRecord) PublicRecord(refTime time.Time) (MDTCarcassRecord, bool) {
	if r.IsSensitive && r.InQuarantineAt(refTime) {
		return MDTCarcassRecord{}, false
	}
	pub := r
	pub.RawLatitude = 0
	pub.RawLongitude = 0
	pub.SMSX = 0
	pub.SMSY = 0
	pub.SMS_X_CORD = 0
	pub.SMS_Y_CORD = 0
	pub.InQuarantine = false
	return pub, true
}

// MarshalJSON emits the public record. Precise state-plane and raw WGS84
// coordinates are never encoded. Sensitive locations still inside the 30-day
// delay are omitted entirely.
func (r MDTCarcassRecord) MarshalJSON() ([]byte, error) {
	type publicRecord struct {
		CollectionDate    time.Time `json:"collection_date"`
		Route             string    `json:"route"`
		Milepost          float64   `json:"milepost"`
		AnimalType        string    `json:"animal_type"`
		NormalizedSpecies string    `json:"normalized_species"`
		Gender            string    `json:"gender"`
		Quantity          int       `json:"quantity"`
		Longitude         float64   `json:"longitude,omitempty"`
		Latitude          float64   `json:"latitude,omitempty"`
		IsSensitive       bool      `json:"is_sensitive"`
		H3R6              string    `json:"h3_r6,omitempty"`
		H3Index           string    `json:"h3_index,omitempty"`
		QuarantineUntil   time.Time `json:"quarantine_until,omitempty"`
		InQuarantine      bool      `json:"in_quarantine"`
	}
	out := publicRecord{
		CollectionDate:    r.CollectionDate,
		Route:             r.Route,
		Milepost:          r.Milepost,
		AnimalType:        r.AnimalType,
		NormalizedSpecies: r.NormalizedSpecies,
		Gender:            r.Gender,
		Quantity:          r.Quantity,
		IsSensitive:       r.IsSensitive,
		QuarantineUntil:   r.QuarantineUntil,
		InQuarantine:      r.InQuarantine,
	}
	if !r.IsSensitive || !r.InQuarantine {
		out.Longitude = r.Longitude
		out.Latitude = r.Latitude
		out.H3R6 = r.H3R6
		out.H3Index = r.H3Index
	}
	return json.Marshal(out)
}

// Cell returns the H3 resolution 6 cell identifier.
func (r MDTCarcassRecord) Cell() string {
	return r.H3R6
}

// Coords returns the effective published coordinates (generalized if sensitive).
func (r MDTCarcassRecord) Coords() (lat, lon float64) {
	return r.Latitude, r.Longitude
}

// RawCoords returns the ungeneralized native GPS coordinates.
func (r MDTCarcassRecord) RawCoords() (lat, lon float64) {
	return r.RawLatitude, r.RawLongitude
}

// Supported date parsing layouts for MDT CSV exports.
var mdtDateLayouts = []string{
	"2006-01-02",
	"01/02/2006",
	"1/2/2006",
	"01/02/06",
	"1/2/06",
	"2006/01/02",
	"01/02/2006 15:04:05",
	"1/2/2006 15:04:05",
	"01/02/2006 03:04:05 PM",
	"1/2/2006 03:04:05 PM",
	"01/02/2006 3:04:05 PM",
	"1/2/2006 3:04:05 PM",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	time.RFC3339,
}

// ParseMDTDate parses a date string across various Excel and CSV date representations.
func ParseMDTDate(s string) (time.Time, error) {
	clean := strings.TrimSpace(s)
	if clean == "" {
		return time.Time{}, errors.New("empty date string")
	}

	for _, layout := range mdtDateLayouts {
		if t, err := time.Parse(layout, clean); err == nil {
			return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC), nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized date format: %q", s)
}

// NormalizeMDTSpecies standardizes common animal names found in MDT records
// into canonical snake_case taxon identifiers.
func NormalizeMDTSpecies(raw string) string {
	clean := strings.ToLower(strings.TrimSpace(raw))
	if clean == "" {
		return "unknown"
	}

	// Canonical mapping table for Montana fauna
	switch clean {
	case "white-tailed deer", "white tailed deer", "white tail deer", "white-tail deer",
		"whitetail deer", "whitetail", "wt deer", "wt_deer", "whitetailed deer":
		return "white_tailed_deer"

	case "mule deer", "mule_deer", "md deer", "md", "black-tailed deer", "blacktailed deer":
		return "mule_deer"

	case "deer", "unknown deer", "unspecified deer":
		return "deer"

	case "elk", "wapiti", "rocky mountain elk":
		return "elk"

	case "moose", "alces alces", "shiras moose":
		return "moose"

	case "pronghorn", "antelope", "pronghorn antelope":
		return "pronghorn"

	case "bighorn sheep", "bighorn", "mountain sheep", "rocky mountain bighorn sheep":
		return "bighorn_sheep"

	case "mountain goat", "rocky mountain goat":
		return "mountain_goat"

	case "black bear", "american black bear":
		return "black_bear"

	case "grizzly bear", "grizzly", "silvertip", "ursus arctos", "ursus arctos horribilis":
		return "grizzly_bear"

	case "canada lynx", "lynx", "lynx canadensis":
		return "lynx"

	case "wolf", "gray wolf", "grey wolf", "timber wolf", "canis lupus":
		return "gray_wolf"

	case "coyote", "canis latrans":
		return "coyote"

	case "mountain lion", "cougar", "puma", "panther", "catamount", "felis concolor", "puma concolor":
		return "mountain_lion"

	case "bobcat", "lynx rufus":
		return "bobcat"

	case "wolverine", "gulo gulo", "gulo gulo luscus":
		return "wolverine"

	case "red fox", "fox", "vulpes vulpes":
		return "red_fox"

	case "raccoon", "procyon lotor":
		return "raccoon"

	case "badger", "american badger", "taxidea taxus":
		return "badger"

	case "striped skunk", "skunk", "mephitis mephitis":
		return "skunk"

	case "porcupine", "north american porcupine", "erethizon dorsatum":
		return "porcupine"

	case "wild turkey", "turkey", "meleagris gallopavo":
		return "wild_turkey"

	case "river otter", "otter":
		return "river_otter"

	case "beaver", "castor canadensis":
		return "beaver"

	case "bison", "buffalo", "american bison":
		return "bison"

	case "unknown", "other", "n/a", "na", "unidentified":
		return "unknown"
	}

	// Fallback heuristic normalization: lowercase, replace spaces and dashes with underscores
	var b strings.Builder
	lastUnderscore := false
	for _, r := range clean {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastUnderscore = false
		} else if r == ' ' || r == '-' || r == '_' || r == '/' {
			if !lastUnderscore && b.Len() > 0 {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
	}
	result := strings.TrimRight(b.String(), "_")
	if result == "" {
		return "unknown"
	}
	return result
}

// IsMDTSensitiveSpecies determines if a normalized or raw species name
// represents a sensitive, threatened, or endangered taxon requiring public
// spatial generalization (to H3 resolution 6) and 30-day quarantine delay.
// Specifically covers Grizzly Bear, Canada Lynx, Gray Wolf, and Wolverine.
func IsMDTSensitiveSpecies(species string) bool {
	norm := NormalizeMDTSpecies(species)
	switch norm {
	case "grizzly_bear", "lynx", "gray_wolf", "wolverine":
		return true
	default:
		return false
	}
}

// NormalizeMDTGender standardizes reported animal gender values.
func NormalizeMDTGender(raw string) string {
	clean := strings.ToLower(strings.TrimSpace(raw))
	switch clean {
	case "m", "male", "buck", "bull", "ram", "boar":
		return "male"
	case "f", "female", "doe", "cow", "ewe", "sow":
		return "female"
	default:
		return "unknown"
	}
}

// ParseMDT parses Montana Department of Transportation carcass salvage records.
func ParseMDT(r io.Reader) ([]MDTCarcassRecord, error) {
	return ParseMDTCarcassCSV(r)
}

// ParseMDTCarcassCSV parses MDT carcass salvage records from a CSV reader.
// It performs:
//  1. Header identification for `Collection Date`, `Route`, `Milepost`, `Animal Type`, `Gender`, `Quantity`, `SMS_X_CORD`, `SMS_Y_CORD`.
//  2. Pure Go Lambert Conformal Conic coordinate transformation from Montana State Plane NAD83 (EPSG:32100) to WGS84 (EPSG:4326).
//  3. Species normalization into canonical snake_case taxon names.
//  4. Sensitive species policy enforcement: grizzly bear, Canada lynx, gray wolf, and wolverine
//     are generalized to Uber H3 resolution 6 centroids and withheld from public JSON for 30 days.
func ParseMDTCarcassCSV(r io.Reader) ([]MDTCarcassRecord, error) {
	if r == nil {
		return nil, errors.New("nil CSV reader")
	}

	csvReader := csv.NewReader(r)
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1 // Allow variable row lengths if trailing empty cells exist

	rawHeader, err := csvReader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("empty CSV input")
		}
		return nil, fmt.Errorf("read CSV header: %w", err)
	}

	headerMap := map[string]int{}
	for i, h := range rawHeader {
		// Strip UTF-8 BOM if present on the first cell
		clean := strings.TrimPrefix(h, "\uFEFF")
		clean = strings.TrimPrefix(clean, "\xef\xbb\xbf")
		clean = strings.ToLower(strings.TrimSpace(clean))
		clean = strings.ReplaceAll(clean, "_", " ")
		clean = strings.ReplaceAll(clean, "-", " ")
		headerMap[clean] = i
	}

	// Helper to find column index matching aliases
	findCol := func(aliases ...string) int {
		for _, alias := range aliases {
			target := strings.ToLower(alias)
			target = strings.ReplaceAll(target, "_", " ")
			target = strings.ReplaceAll(target, "-", " ")
			if idx, ok := headerMap[target]; ok {
				return idx
			}
		}
		return -1
	}

	colDate := findCol("collection date", "collectiondate", "date", "pickup date", "salvage date")
	colRoute := findCol("route", "route id", "routeid", "hwy", "highway", "road", "corridor")
	colMilepost := findCol("milepost", "mp", "mile post", "beg milepost")
	colAnimal := findCol("animal type", "animaltype", "animal", "species", "species name")
	colGender := findCol("gender", "sex")
	colQuantity := findCol("quantity", "qty", "count", "number of roadkill", "total")
	colX := findCol("sms x cord", "sms x coord", "sms x", "sms x cord (m)", "x cord", "x coord", "easting", "x")
	colY := findCol("sms y cord", "sms y coord", "sms y", "sms y cord (m)", "y cord", "y coord", "northing", "y")

	// Verify required columns
	var missingCols []string
	if colDate < 0 {
		missingCols = append(missingCols, "Collection Date")
	}
	if colRoute < 0 {
		missingCols = append(missingCols, "Route")
	}
	if colAnimal < 0 {
		missingCols = append(missingCols, "Animal Type")
	}
	if colX < 0 {
		missingCols = append(missingCols, "SMS_X_CORD")
	}
	if colY < 0 {
		missingCols = append(missingCols, "SMS_Y_CORD")
	}

	if len(missingCols) > 0 {
		return nil, fmt.Errorf("missing required columns in MDT carcass CSV: %s", strings.Join(missingCols, ", "))
	}

	var records []MDTCarcassRecord
	lineNum := 1

	for {
		row, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		lineNum++
		if err != nil {
			return nil, fmt.Errorf("read CSV row %d: %w", lineNum, err)
		}

		// Skip entirely blank rows
		isBlank := true
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				isBlank = false
				break
			}
		}
		if isBlank {
			continue
		}

		get := func(colIdx int) string {
			if colIdx >= 0 && colIdx < len(row) {
				return strings.TrimSpace(row[colIdx])
			}
			return ""
		}

		// Parse collection date
		dateStr := get(colDate)
		collectionDate, err := ParseMDTDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid Collection Date %q: %w", lineNum, dateStr, err)
		}

		route := NormalizeMDTRoute(get(colRoute))

		// Milepost (optional, defaults to 0.0)
		var milepost float64
		if colMilepost >= 0 {
			mpStr := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(get(colMilepost), "MP"), "mp"))
			if mpStr != "" && mpStr != "-" && !strings.EqualFold(mpStr, "N/A") {
				if mp, err := strconv.ParseFloat(mpStr, 64); err == nil && !math.IsNaN(mp) && !math.IsInf(mp, 0) {
					milepost = mp
				}
			}
		}

		// Animal type and normalized species
		animalType := get(colAnimal)
		if animalType == "" {
			animalType = "Unknown"
		}
		normalizedSpecies := NormalizeMDTSpecies(animalType)

		// Gender
		gender := "unknown"
		if colGender >= 0 {
			gender = NormalizeMDTGender(get(colGender))
		}

		// Quantity (defaults to 1)
		quantity := 1
		if colQuantity >= 0 {
			qStr := get(colQuantity)
			if qStr != "" {
				if q, err := strconv.Atoi(qStr); err == nil && q > 0 {
					quantity = q
				}
			}
		}

		// Coordinates: SMS_X_CORD (Easting) and SMS_Y_CORD (Northing) in meters
		xStr := get(colX)
		yStr := get(colY)
		easting, err := strconv.ParseFloat(strings.ReplaceAll(xStr, ",", ""), 64)
		if err != nil || math.IsNaN(easting) || math.IsInf(easting, 0) {
			return nil, fmt.Errorf("row %d: invalid SMS_X_CORD %q", lineNum, xStr)
		}
		northing, err := strconv.ParseFloat(strings.ReplaceAll(yStr, ",", ""), 64)
		if err != nil || math.IsNaN(northing) || math.IsInf(northing, 0) {
			return nil, fmt.Errorf("row %d: invalid SMS_Y_CORD %q", lineNum, yStr)
		}

		// Transform Montana State Plane (EPSG:32100) -> WGS84 (EPSG:4326)
		rawLat, rawLon, err := geo.StatePlane32100ToWGS84(easting, northing)
		if err != nil {
			return nil, fmt.Errorf("row %d: state plane coordinate transformation: %w", lineNum, err)
		}

		// Check sensitive species policy
		isSensitive := IsMDTSensitiveSpecies(normalizedSpecies)

		// Compute H3 resolution 6 cell index
		h3CellStr, centLat, centLon, err := ComputeH3R6(rawLat, rawLon)
		if err != nil {
			return nil, fmt.Errorf("row %d: calculate H3 resolution 6 cell: %w", lineNum, err)
		}

		rec := MDTCarcassRecord{
			CollectionDate:    collectionDate,
			Route:             route,
			Milepost:          milepost,
			AnimalType:        animalType,
			NormalizedSpecies: normalizedSpecies,
			Gender:            gender,
			Quantity:          quantity,
			SMSX:              easting,
			SMSY:              northing,
			SMS_X_CORD:        easting,
			SMS_Y_CORD:        northing,
			RawLatitude:       rawLat,
			RawLongitude:      rawLon,
			IsSensitive:       isSensitive,
			H3R6:              h3CellStr,
			H3Index:           h3CellStr,
		}

		if isSensitive {
			rec.Latitude = centLat
			rec.Longitude = centLon
			rec.QuarantineUntil = collectionDate.Add(SensitiveQuarantineDuration)
			rec.InQuarantine = rec.InQuarantineAt(time.Now().UTC())
		} else {
			rec.Latitude = rawLat
			rec.Longitude = rawLon
		}

		records = append(records, rec)
	}

	return records, nil
}

// ParseMDTCarcassBytes parses MDT carcass records from raw bytes.
func ParseMDTCarcassBytes(data []byte) ([]MDTCarcassRecord, error) {
	return ParseMDTCarcassCSV(bytes.NewReader(data))
}
