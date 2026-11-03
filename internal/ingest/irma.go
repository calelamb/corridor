// Package ingest normalizes licensed source artifacts without publishing raw records.
package ingest

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	h3 "github.com/akhenakh/goh3"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultH3Resolution specifies H3 resolution 6 (~36 km² cell area) for sensitive species public generalization.
	DefaultH3Resolution = 6
	// SensitiveQuarantineDuration defines the mandatory 30-day temporal delay for sensitive species.
	SensitiveQuarantineDuration = 30 * 24 * time.Hour
)

// IRMAArtifact returns catalog metadata for NPS IRMA Reference 2317362.
// The WHMMO CSV is not in this repository and is not listed in data/SHA256SUMS.
// SHA256 and RetrievedAt stay empty until that file is actually acquired and hashed.
func IRMAArtifact() Artifact {
	return Artifact{
		SourceID:  "23173620-0000-4000-8000-000000231736",
		Name:      "NPS IRMA 2317362 · Wildlife Mortality and Morbidity Observations (WHMMO)",
		URL:       "https://irma.nps.gov/DataStore/Reference/Profile/2317362",
		License:   "Public Domain · U.S. National Park Service",
		SHA256:    "",
		ObjectKey: "irma/2317362/WHMMO_data.csv",
	}
}

// IRMARecord represents an observation parsed from NPS IRMA Reference 2317362 (WHMMO_data.csv).
type IRMARecord struct {
	RawID             string    `json:"raw_id"`
	ParkCode          string    `json:"park_code"`
	ObservationDate   time.Time `json:"observation_date"`
	DatePrecision     string    `json:"date_precision"` // "day", "month", or "year"
	CommonName        string    `json:"common_name"`
	ScientificName    string    `json:"scientific_name"`
	NormalizedSpecies string    `json:"normalized_species"`
	Road              string    `json:"road"`
	Milepost          string    `json:"milepost"`
	Latitude          float64   `json:"latitude"`
	Longitude         float64   `json:"longitude"`
	HasCoordinates    bool      `json:"has_coordinates"`
	CauseOfDeath      string    `json:"cause_of_death"`
	IsRoadkill        bool      `json:"is_roadkill"`
	IsSensitive       bool      `json:"is_sensitive"`
	H3R6              string    `json:"h3_r6,omitempty"`
	GeneralizedLat    float64   `json:"generalized_lat,omitempty"`
	GeneralizedLon    float64   `json:"generalized_lon,omitempty"`
	QuarantineUntil   time.Time `json:"quarantine_until,omitempty"`
	InQuarantine      bool      `json:"in_quarantine"`
	Comments          string    `json:"comments,omitempty"`
}

// IRMAQA tracks ingestion quality metrics for NPS IRMA observations.
type IRMAQA struct {
	Read           int64            `json:"read"`
	Accepted       int64            `json:"accepted"`
	Rejected       int64            `json:"rejected"`
	Duplicates     int64            `json:"duplicates"`
	OutsideScope   int64            `json:"outside_scope"`
	SensitiveCount int64            `json:"sensitive_count"`
	RoadkillCount  int64            `json:"roadkill_count"`
	MissingCoords  int64            `json:"missing_coords"`
	Reasons        map[string]int64 `json:"reasons"`
}

// IRMADataset groups parsed records and execution quality metrics.
type IRMADataset struct {
	Records []IRMARecord `json:"records"`
	QA      IRMAQA       `json:"qa"`
}

// IRMAFilter provides criteria to subset IRMA records during parsing.
type IRMAFilter struct {
	ParkCodes          []string   `json:"park_codes,omitempty"`          // e.g. ["YOSE", "YELL"]
	StartDate          *time.Time `json:"start_date,omitempty"`          // records on or after StartDate
	EndDate            *time.Time `json:"end_date,omitempty"`            // records on or before EndDate
	MortalityCause     string     `json:"mortality_cause,omitempty"`     // filter cause substring
	OnlyRoadkill       bool       `json:"only_roadkill,omitempty"`       // restrict to vehicle strikes
	RequireCoordinates bool       `json:"require_coordinates,omitempty"` // reject rows missing coords
	ReferenceTime      *time.Time `json:"reference_time,omitempty"`      // reference time for quarantine check
}

// InQuarantineAt evaluates whether the record is under sensitive quarantine at refTime.
func (r IRMARecord) InQuarantineAt(refTime time.Time) bool {
	if !r.IsSensitive {
		return false
	}
	return refTime.Before(r.QuarantineUntil)
}

// PublicRecord produces a privacy-compliant record for public feeds.
// If sensitive and currently quarantined, it returns ok=false.
// If sensitive and released, coordinates are generalized to the H3-r6 cell centroid.
func (r IRMARecord) PublicRecord(refTime time.Time) (IRMARecord, bool) {
	if !r.IsSensitive {
		return r, true
	}
	if r.InQuarantineAt(refTime) {
		return IRMARecord{}, false
	}
	pub := r
	if r.HasCoordinates {
		pub.Latitude = r.GeneralizedLat
		pub.Longitude = r.GeneralizedLon
	}
	return pub, true
}

// ToRecord converts an IRMARecord into Corridor's canonical ingest.Record.
func (r IRMARecord) ToRecord() Record {
	unc := 50.0
	if r.IsSensitive {
		unc = 6000.0 // ~H3-r6 bounding radius
	}
	if !r.HasCoordinates {
		unc = 50000.0
	}
	sp := r.NormalizedSpecies
	if r.ScientificName != "" {
		sp = r.ScientificName
	}
	end := r.ObservationDate.AddDate(0, 0, 1)
	switch r.DatePrecision {
	case "month":
		end = r.ObservationDate.AddDate(0, 1, 0)
	case "year":
		end = r.ObservationDate.AddDate(1, 0, 0)
	}

	lon, lat := r.Longitude, r.Latitude
	if r.IsSensitive && r.HasCoordinates {
		lon, lat = r.GeneralizedLon, r.GeneralizedLat
	}
	return Record{
		ID:          r.RawID,
		Species:     sp,
		Road:        r.Road,
		Survey:      fmt.Sprintf("NPS-IRMA-%s", r.ParkCode),
		Reference:   fmt.Sprintf("IRMA:2317362:%s:%s", r.ParkCode, r.RawID),
		Precision:   r.DatePrecision,
		Longitude:   lon,
		Latitude:    lat,
		Uncertainty: unc,
		Start:       r.ObservationDate,
		End:         end,
		Quantity:    1,
	}
}

// MatchesPark checks if a record's park code passes the filter.
func (f *IRMAFilter) MatchesPark(parkCode string) bool {
	if f == nil || len(f.ParkCodes) == 0 {
		return true
	}
	p := strings.ToUpper(strings.TrimSpace(parkCode))
	for _, code := range f.ParkCodes {
		if strings.ToUpper(strings.TrimSpace(code)) == p {
			return true
		}
	}
	return false
}

// MatchesDate checks if an observation date falls within the filter window.
func (f *IRMAFilter) MatchesDate(t time.Time) bool {
	if f == nil {
		return true
	}
	if f.StartDate != nil && t.Before(*f.StartDate) {
		return false
	}
	if f.EndDate != nil && t.After(*f.EndDate) {
		return false
	}
	return true
}

// MatchesCause verifies if the mortality cause matches filter requirements.
func (f *IRMAFilter) MatchesCause(cause string, isRoadkill bool) bool {
	if f == nil {
		return true
	}
	if f.OnlyRoadkill && !isRoadkill {
		return false
	}
	if f.MortalityCause != "" {
		target := strings.ToLower(strings.TrimSpace(f.MortalityCause))
		if !strings.Contains(strings.ToLower(cause), target) {
			return false
		}
	}
	return true
}

// ParseIRMA parses a CSV stream from NPS IRMA Reference 2317362 (WHMMO_data.csv).
func ParseIRMA(input io.Reader, filter *IRMAFilter) (IRMADataset, error) {
	reader := csv.NewReader(io.LimitReader(input, 256<<20))
	header, err := reader.Read()
	if err != nil {
		return IRMADataset{}, fmt.Errorf("read CSV header: %w", err)
	}

	indices, err := resolveIRMAColumns(header)
	if err != nil {
		return IRMADataset{}, err
	}

	now := time.Now().UTC()
	if filter != nil && filter.ReferenceTime != nil {
		now = *filter.ReferenceTime
	}

	taxonomy := getTaxonomyRegistry()
	result := IRMADataset{
		Records: make([]IRMARecord, 0),
		QA: IRMAQA{
			Reasons: make(map[string]int64),
		},
	}
	seen := make(map[string]bool)

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return IRMADataset{}, fmt.Errorf("read CSV row %d: %w", result.QA.Read+1, err)
		}
		result.QA.Read++
		if result.QA.Read > 500000 {
			return IRMADataset{}, errors.New("source row limit exceeded")
		}

		record, err := parseIRMARow(row, indices, taxonomy, now)
		if err != nil {
			result.QA.Rejected++
			result.QA.Reasons[err.Error()]++
			if strings.Contains(err.Error(), "coordinates") {
				result.QA.MissingCoords++
			}
			continue
		}

		if !record.HasCoordinates {
			result.QA.MissingCoords++
			if filter != nil && filter.RequireCoordinates {
				result.QA.Rejected++
				result.QA.Reasons["missing coordinates"]++
				continue
			}
		}

		// Apply filters
		if !filter.MatchesPark(record.ParkCode) {
			result.QA.OutsideScope++
			continue
		}
		if !filter.MatchesDate(record.ObservationDate) {
			result.QA.OutsideScope++
			continue
		}
		if !filter.MatchesCause(record.CauseOfDeath, record.IsRoadkill) {
			result.QA.OutsideScope++
			continue
		}

		// Duplicate suppression
		if seen[record.RawID] {
			result.QA.Duplicates++
			continue
		}
		seen[record.RawID] = true

		if record.IsSensitive {
			result.QA.SensitiveCount++
		}
		if record.IsRoadkill {
			result.QA.RoadkillCount++
		}

		result.Records = append(result.Records, record)
		result.QA.Accepted++
	}

	return result, nil
}

type irmaColumnIndices struct {
	ParkCode     int
	Date         int
	Species      int
	Road         int
	Milepost     int
	Lat          int
	Lon          int
	CauseOfDeath int
	ID           int
	Comments     int
}

func resolveIRMAColumns(header []string) (irmaColumnIndices, error) {
	idx := irmaColumnIndices{
		ParkCode:     -1,
		Date:         -1,
		Species:      -1,
		Road:         -1,
		Milepost:     -1,
		Lat:          -1,
		Lon:          -1,
		CauseOfDeath: -1,
		ID:           -1,
		Comments:     -1,
	}

	for i, h := range header {
		norm := strings.ToLower(strings.TrimSpace(h))
		normClean := strings.ReplaceAll(strings.ReplaceAll(norm, "_", ""), " ", "")
		normClean = strings.ReplaceAll(normClean, "/", "")

		switch {
		case matchesCol(normClean, "parkcode", "unitcode", "park", "parkunit"):
			if idx.ParkCode == -1 {
				idx.ParkCode = i
			}
		case matchesCol(normClean, "observationdate", "obsdate", "date", "eventdate", "observedat"):
			if idx.Date == -1 {
				idx.Date = i
			}
		case matchesCol(normClean, "commonname", "species", "scientificname", "animal", "speciesname", "wildlife"):
			if idx.Species == -1 {
				idx.Species = i
			}
		case matchesCol(normClean, "roadroutename", "roadname", "road", "route", "routename", "highway"):
			if idx.Road == -1 {
				idx.Road = i
			}
		case matchesCol(normClean, "milepostlocationdescription", "locationdescription", "milepost", "location", "mp"):
			if idx.Milepost == -1 {
				idx.Milepost = i
			}
		case matchesCol(normClean, "latitude", "lat", "decimallatitude", "y"):
			if idx.Lat == -1 {
				idx.Lat = i
			}
		case matchesCol(normClean, "longitude", "long", "lon", "decimallongitude", "x"):
			if idx.Lon == -1 {
				idx.Lon = i
			}
		case matchesCol(normClean, "causeofdeath", "mortalitytype", "mortalitycause", "cause", "deathcause"):
			if idx.CauseOfDeath == -1 {
				idx.CauseOfDeath = i
			}
		case matchesCol(normClean, "observationid", "recordid", "id", "eventid", "objectid"):
			if idx.ID == -1 {
				idx.ID = i
			}
		case matchesCol(normClean, "comments", "comment", "notes", "remarks", "description"):
			if idx.Comments == -1 {
				idx.Comments = i
			}
		}
	}

	if idx.ParkCode == -1 {
		return idx, errors.New("missing required column: Park Code")
	}
	if idx.Date == -1 {
		return idx, errors.New("missing required column: Observation Date")
	}
	if idx.Species == -1 {
		return idx, errors.New("missing required column: Species / Common Name")
	}

	return idx, nil
}

func matchesCol(val string, aliases ...string) bool {
	for _, a := range aliases {
		if val == a {
			return true
		}
	}
	return false
}

func parseIRMARow(row []string, idx irmaColumnIndices, taxonomy *TaxonomyRegistry, now time.Time) (IRMARecord, error) {
	get := func(i int) string {
		if i >= 0 && i < len(row) {
			return strings.TrimSpace(row[i])
		}
		return ""
	}

	parkCode := strings.ToUpper(get(idx.ParkCode))
	if parkCode == "" {
		return IRMARecord{}, errors.New("missing park code")
	}

	dateStr := get(idx.Date)
	obsDate, precision, err := parseIRMADate(dateStr)
	if err != nil {
		return IRMARecord{}, err
	}

	rawSpecies := get(idx.Species)
	if rawSpecies == "" {
		return IRMARecord{}, errors.New("empty species")
	}

	speciesInfo := taxonomy.Normalize(rawSpecies, parkCode)

	road := get(idx.Road)
	milepost := get(idx.Milepost)
	cause := get(idx.CauseOfDeath)
	comments := get(idx.Comments)
	rawID := get(idx.ID)

	if rawID == "" {
		hash := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%s|%s|%s", parkCode, dateStr, rawSpecies, road, milepost, cause)))
		rawID = fmt.Sprintf("nps-%s-%s-%s", strings.ToLower(parkCode), obsDate.Format("20060102"), hex.EncodeToString(hash[:4]))
	}

	rec := IRMARecord{
		RawID:             rawID,
		ParkCode:          parkCode,
		ObservationDate:   obsDate,
		DatePrecision:     precision,
		CommonName:        speciesInfo.CommonName,
		ScientificName:    speciesInfo.ScientificName,
		NormalizedSpecies: speciesInfo.NormalizedName,
		Road:              road,
		Milepost:          milepost,
		CauseOfDeath:      cause,
		IsRoadkill:        IsVehicleStrike(cause),
		IsSensitive:       speciesInfo.IsSensitive,
		Comments:          comments,
	}

	latStr := get(idx.Lat)
	lonStr := get(idx.Lon)

	if latStr != "" && lonStr != "" {
		lat, errLat := strconv.ParseFloat(latStr, 64)
		lon, errLon := strconv.ParseFloat(lonStr, 64)
		if errLat != nil || errLon != nil || math.IsNaN(lat) || math.IsNaN(lon) || math.IsInf(lat, 0) || math.IsInf(lon, 0) {
			return IRMARecord{}, errors.New("invalid coordinates")
		}
		if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
			return IRMARecord{}, errors.New("invalid coordinates out of bounds")
		}
		rec.Latitude = lat
		rec.Longitude = lon
		rec.HasCoordinates = true

		// Assign H3 resolution 6 cell
		cellStr, cLat, cLon, err := ComputeH3R6(lat, lon)
		if err != nil {
			return IRMARecord{}, fmt.Errorf("assign h3 cell: %w", err)
		}
		rec.H3R6 = cellStr
		rec.GeneralizedLat = cLat
		rec.GeneralizedLon = cLon
	} else {
		rec.HasCoordinates = false
	}

	// Sensitive species enforcement
	if rec.IsSensitive {
		rec.QuarantineUntil = obsDate.Add(SensitiveQuarantineDuration)
		rec.InQuarantine = rec.InQuarantineAt(now)
	}

	return rec, nil
}

// ParseH3Cell parses an H3 hex string into an H3 Cell.
func ParseH3Cell(s string) (h3.Cell, error) {
	val, err := strconv.ParseUint(strings.TrimSpace(s), 16, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid h3 hex: %w", err)
	}
	cell := h3.Cell(val)
	if !cell.IsValid() {
		return 0, errors.New("invalid h3 cell")
	}
	return cell, nil
}

// ComputeH3R6 calculates the Uber H3 resolution 6 hexagon index and its centroid.
func ComputeH3R6(lat, lon float64) (cellStr string, centroidLat, centroidLon float64, err error) {
	cell := h3.LatLngToCell(h3.NewLatLng(lat, lon), DefaultH3Resolution)
	if !cell.IsValid() {
		return "", 0, 0, errors.New("invalid h3 cell generated")
	}
	centroid := cell.LatLng()
	return cell.String(), centroid.Lat, centroid.Lng, nil
}

// IsVehicleStrike identifies if the mortality reason was a motor vehicle collision.
func IsVehicleStrike(cause string) bool {
	c := strings.ToLower(strings.TrimSpace(cause))
	if c == "" {
		return false
	}
	keywords := []string{
		"vehicle", "strike", "roadkill", "collision", "hit by", "struck by",
		"car", "truck", "automobile", "traffic", "mva", "road", "run over",
	}
	for _, kw := range keywords {
		if strings.Contains(c, kw) {
			return true
		}
	}
	return false
}

func parseIRMADate(raw string) (time.Time, string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return time.Time{}, "", errors.New("empty observation date")
	}

	formats := []struct {
		layout    string
		precision string
	}{
		{time.RFC3339, "day"},
		{"2006-01-02 15:04:05", "day"},
		{"2006-01-02 15:04", "day"},
		{"2006-01-02", "day"},
		{"01/02/2006 15:04:05", "day"},
		{"01/02/2006 15:04", "day"},
		{"01/02/2006", "day"},
		{"1/2/2006", "day"},
		{"01-02-2006", "day"},
		{"2006/01/02", "day"},
		{"2006-01", "month"},
		{"01/2006", "month"},
		{"2006", "year"},
	}

	for _, f := range formats {
		if t, err := time.Parse(f.layout, s); err == nil {
			if t.Year() < 1900 || t.Year() > time.Now().Year()+1 {
				return time.Time{}, "", errors.New("observation date year out of range")
			}
			return t.UTC(), f.precision, nil
		}
	}

	return time.Time{}, "", errors.New("unrecognized observation date format")
}

// SpeciesTaxon encapsulates normalized species data and sensitivity attributes.
type SpeciesTaxon struct {
	CommonName     string
	ScientificName string
	NormalizedName string
	IsSensitive    bool
	Status         string
	Category       string
	Aliases        []string
}

// TaxonomyRegistry manages species taxonomic normalization and sensitivity flags.
type TaxonomyRegistry struct {
	lookup map[string]SpeciesTaxon
}

var defaultRegistry *TaxonomyRegistry

func getTaxonomyRegistry() *TaxonomyRegistry {
	if defaultRegistry == nil {
		defaultRegistry = buildTaxonomyRegistry()
	}
	return defaultRegistry
}

func normalizeTaxonKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	parts := strings.Fields(s)
	return strings.Join(parts, " ")
}

func buildTaxonomyRegistry() *TaxonomyRegistry {
	reg := &TaxonomyRegistry{
		lookup: make(map[string]SpeciesTaxon),
	}

	// 1. Load sensitive species definitions
	sensitiveSpecies := []SpeciesTaxon{
		{
			CommonName:     "Sierra Nevada bighorn sheep",
			ScientificName: "Ovis canadensis sierrae",
			NormalizedName: "sierra nevada bighorn sheep",
			IsSensitive:    true,
			Status:         "endangered",
			Category:       "listed_ungulate",
			Aliases:        []string{"sierra bighorn sheep", "bighorn sheep", "sierra nevada bighorn", "ovis canadensis sierrae"},
		},
		{
			CommonName:     "Pacific fisher",
			ScientificName: "Pekania pennanti",
			NormalizedName: "pacific fisher",
			IsSensitive:    true,
			Status:         "endangered",
			Category:       "rare_carnivore",
			Aliases:        []string{"fisher", "pacific fisher", "pekania pennanti", "martes pennanti"},
		},
		{
			CommonName:     "California black bear",
			ScientificName: "Ursus americanus californiensis",
			NormalizedName: "california black bear",
			IsSensitive:    true,
			Status:         "protected",
			Category:       "apex_predator",
			Aliases:        []string{"black bear", "american black bear", "california black bear", "ursus americanus"},
		},
		{
			CommonName:     "Grizzly bear",
			ScientificName: "Ursus arctos horribilis",
			NormalizedName: "grizzly bear",
			IsSensitive:    true,
			Status:         "threatened",
			Category:       "apex_predator",
			Aliases:        []string{"grizzly", "brown bear", "ursus arctos", "ursus arctos horribilis"},
		},
		{
			CommonName:     "Canada lynx",
			ScientificName: "Lynx canadensis",
			NormalizedName: "canada lynx",
			IsSensitive:    true,
			Status:         "threatened",
			Category:       "rare_carnivore",
			Aliases:        []string{"lynx", "canada lynx", "lynx canadensis"},
		},
		{
			CommonName:     "Gray wolf",
			ScientificName: "Canis lupus",
			NormalizedName: "gray wolf",
			IsSensitive:    true,
			Status:         "endangered",
			Category:       "apex_predator",
			Aliases:        []string{"grey wolf", "wolf", "timber wolf", "canis lupus"},
		},
		{
			CommonName:     "Wolverine",
			ScientificName: "Gulo gulo luscus",
			NormalizedName: "wolverine",
			IsSensitive:    true,
			Status:         "threatened",
			Category:       "rare_carnivore",
			Aliases:        []string{"wolverine", "gulo gulo", "gulo gulo luscus"},
		},
	}

	for _, s := range sensitiveSpecies {
		reg.registerTaxon(s)
	}

	// 2. Standard non-sensitive common wildlife
	standardSpecies := []SpeciesTaxon{
		{
			CommonName:     "Mule deer",
			ScientificName: "Odocoileus hemionus",
			NormalizedName: "mule deer",
			Aliases:        []string{"black-tailed deer", "mule deer", "odocoileus hemionus"},
		},
		{
			CommonName:     "White-tailed deer",
			ScientificName: "Odocoileus virginianus",
			NormalizedName: "white-tailed deer",
			Aliases:        []string{"white tailed deer", "whitetailed deer", "odocoileus virginianus"},
		},
		{
			CommonName:     "Rocky Mountain elk",
			ScientificName: "Cervus canadensis",
			NormalizedName: "elk",
			Aliases:        []string{"elk", "wapiti", "cervus canadensis", "cervus elaphus"},
		},
		{
			CommonName:     "Moose",
			ScientificName: "Alces alces",
			NormalizedName: "moose",
			Aliases:        []string{"shiras moose", "alces alces"},
		},
		{
			CommonName:     "Pronghorn",
			ScientificName: "Antilocapra americana",
			NormalizedName: "pronghorn",
			Aliases:        []string{"pronghorn antelope", "antelope", "antilocapra americana"},
		},
		{
			CommonName:     "American bison",
			ScientificName: "Bison bison",
			NormalizedName: "bison",
			Aliases:        []string{"bison", "buffalo", "plains bison", "bison bison"},
		},
		{
			CommonName:     "Coyote",
			ScientificName: "Canis latrans",
			NormalizedName: "coyote",
			Aliases:        []string{"canis latrans"},
		},
		{
			CommonName:     "Bobcat",
			ScientificName: "Lynx rufus",
			NormalizedName: "bobcat",
			Aliases:        []string{"wildcat", "lynx rufus"},
		},
		{
			CommonName:     "Mountain lion",
			ScientificName: "Puma concolor",
			NormalizedName: "mountain lion",
			Aliases:        []string{"cougar", "puma", "panther", "puma concolor"},
		},
		{
			CommonName:     "Red fox",
			ScientificName: "Vulpes vulpes",
			NormalizedName: "red fox",
			Aliases:        []string{"vulpes vulpes"},
		},
		{
			CommonName:     "American badger",
			ScientificName: "Taxidea taxus",
			NormalizedName: "badger",
			Aliases:        []string{"badger", "taxidea taxus"},
		},
		{
			CommonName:     "Striped skunk",
			ScientificName: "Mephitis mephitis",
			NormalizedName: "striped skunk",
			Aliases:        []string{"skunk", "mephitis mephitis"},
		},
		{
			CommonName:     "Raccoon",
			ScientificName: "Procyon lotor",
			NormalizedName: "raccoon",
			Aliases:        []string{"procyon lotor"},
		},
		{
			CommonName:     "North American porcupine",
			ScientificName: "Erethizon dorsatum",
			NormalizedName: "porcupine",
			Aliases:        []string{"erethizon dorsatum"},
		},
		{
			CommonName:     "Yellow-bellied marmot",
			ScientificName: "Marmota flaviventris",
			NormalizedName: "marmot",
			Aliases:        []string{"yellow-bellied marmot", "rock chuck", "marmota flaviventris"},
		},
	}

	for _, s := range standardSpecies {
		reg.registerTaxon(s)
	}

	// 3. Try enriching from config/sensitive_species.yaml if available
	tryEnrichFromYAML(reg)

	return reg
}

func (r *TaxonomyRegistry) registerTaxon(taxon SpeciesTaxon) {
	keys := []string{
		normalizeTaxonKey(taxon.CommonName),
		normalizeTaxonKey(taxon.ScientificName),
		normalizeTaxonKey(taxon.NormalizedName),
	}
	for _, a := range taxon.Aliases {
		keys = append(keys, normalizeTaxonKey(a))
	}
	for _, k := range keys {
		if k != "" {
			r.lookup[k] = taxon
		}
	}
}

func tryEnrichFromYAML(reg *TaxonomyRegistry) {
	paths := []string{
		"config/sensitive_species.yaml",
		filepath.Join("..", "..", "config", "sensitive_species.yaml"),
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var cfg struct {
			Species []struct {
				CommonName     string   `yaml:"common_name"`
				ScientificName string   `yaml:"scientific_name"`
				Status         string   `yaml:"status"`
				Category       string   `yaml:"category"`
				Aliases        []string `yaml:"aliases"`
			} `yaml:"species"`
		}
		if err := yaml.Unmarshal(data, &cfg); err == nil {
			for _, s := range cfg.Species {
				reg.registerTaxon(SpeciesTaxon{
					CommonName:     s.CommonName,
					ScientificName: s.ScientificName,
					NormalizedName: strings.ToLower(s.CommonName),
					IsSensitive:    true,
					Status:         s.Status,
					Category:       s.Category,
					Aliases:        s.Aliases,
				})
			}
			break
		}
	}
}

// Normalize standardizes species names using standard taxonomy and sensitive classification.
func (r *TaxonomyRegistry) Normalize(rawSpecies, parkCode string) SpeciesTaxon {
	trimmed := strings.TrimSpace(rawSpecies)
	if trimmed == "" {
		return SpeciesTaxon{
			CommonName:     "Unknown",
			NormalizedName: "unknown",
			IsSensitive:    false,
		}
	}

	// Extract candidate sub-terms (handles "Common Name (Scientific Name)" or "Common / Scientific")
	candidates := []string{trimmed}
	if idx := strings.Index(trimmed, "("); idx > 0 {
		common := strings.TrimSpace(trimmed[:idx])
		candidates = append(candidates, common)
		if end := strings.Index(trimmed, ")"); end > idx {
			sci := strings.TrimSpace(trimmed[idx+1 : end])
			candidates = append(candidates, sci)
		}
	}
	if parts := strings.Split(trimmed, "/"); len(parts) > 1 {
		for _, p := range parts {
			candidates = append(candidates, strings.TrimSpace(p))
		}
	}

	for _, c := range candidates {
		k := normalizeTaxonKey(c)
		if taxon, ok := r.lookup[k]; ok {
			return taxon
		}
	}

	// Substring sensitive species heuristics (e.g. "Grizzly bear subadult", "Sierra bighorn ram")
	normFull := normalizeTaxonKey(trimmed)
	for k, taxon := range r.lookup {
		if taxon.IsSensitive && strings.Contains(normFull, k) {
			return taxon
		}
	}

	// Fallback for unlisted/unknown species
	return SpeciesTaxon{
		CommonName:     trimmed,
		ScientificName: "",
		NormalizedName: strings.ToLower(trimmed),
		IsSensitive:    false,
	}
}
