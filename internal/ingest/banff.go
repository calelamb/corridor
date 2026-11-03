// Package ingest normalizes licensed source artifacts without publishing raw records.
package ingest

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

// BanffAttribution is the default attribution sentence from the Open Government
// Licence – Canada, version 2.0.
const BanffAttribution = "Contains information licensed under the Open Government Licence – Canada."

// BanffCrossingCountArtifact returns provenance metadata for the Parks Canada
// Open Government CSV banff_np_wildlife_crossing_structures_2014-2017_data.csv.
//
// SHA256 and ObjectKey stay empty: the CSV has not been acquired into object
// storage in this workspace, so no digest can be recorded. RetrievedAt is the
// time the file URL was requested during the 2026-11-02 endpoint scout
// (docs/research/2026-11-02-canada-banff-scout.md); it is not a checksum.
func BanffCrossingCountArtifact() Artifact {
	return Artifact{
		SourceID:    "d5fd45b9-7c70-4339-a68b-e60d7cbf42e8",
		Name:        "Parks Canada · Wildlife Crossing Structures – Banff 2014-2017",
		URL:         "https://open.canada.ca/data/en/dataset/d5fd45b9-7c70-4339-a68b-e60d7cbf42e8",
		License:     "Open Government Licence – Canada",
		SHA256:      "",
		ObjectKey:   "",
		RetrievedAt: time.Date(2026, 11, 3, 0, 39, 38, 0, time.UTC),
	}
}

// BanffCount is an optional integer count. A blank source cell yields
// Known=false; the digit 0 yields Known=true with Value 0. The two are distinct
// because the publisher left some cells empty rather than writing zero.
type BanffCount struct {
	Value int
	Known bool
}

// MarshalJSON emits null for an unknown count so consumers cannot read a blank
// cell as zero.
func (c BanffCount) MarshalJSON() ([]byte, error) {
	if !c.Known {
		return []byte("null"), nil
	}
	return json.Marshal(c.Value)
}

// BanffCrossingCountRecord is one row of the Banff crossing-structure camera
// summary: counts for one named structure over one image-date span. The source
// has no latitude, longitude, or geometry column, and this record carries none.
type BanffCrossingCountRecord struct {
	Region        string     `json:"region"`
	Location      string     `json:"location"`
	ImageYear     int        `json:"image_year"`
	FirstImage    time.Time  `json:"first_image"`
	LastImage     time.Time  `json:"last_image"`
	HikersRunners BanffCount `json:"hikers_runners"`
	HorseRiders   BanffCount `json:"horse_riders"`
	Bicyclists    BanffCount `json:"bicyclists"`
	Cougars       BanffCount `json:"cougars"`
	Elk           BanffCount `json:"elk"`
	GrizzlyBears  BanffCount `json:"grizzly_bears"`
	Lynx          BanffCount `json:"lynx"`
	Wolves        BanffCount `json:"wolves"`
	Deer          BanffCount `json:"deer"`
}

// banffColumns lists the verified English header names in file order.
var banffColumns = []string{
	"region",
	"location",
	"year images were taken",
	"date of first image",
	"date of last image",
	"number of hikers and runners",
	"number of horse riders",
	"number of bicyclists",
	"number of cougars",
	"number of elk",
	"number of grizzly bears",
	"number of lynx",
	"number of wolves",
	"number of deer",
}

// banffMaxBytes bounds the CSV read; input at or above this size is rejected
// rather than silently truncated.
const banffMaxBytes = 64 << 20

// Image dates in the file are day/month/year. Go's "2"/"1" elements accept one
// or two digits; month-first, ISO, and two-digit-year forms are rejected so a
// swapped day/month can never be read silently.
const banffDateLayout = "2/1/2006"

// ParseBanffCrossingCounts parses the Banff crossing-structure count CSV.
// Input bytes that are not valid UTF-8 are decoded as Windows-1252, the
// encoding observed on the published file. The French label row that follows
// the English header is skipped. Unknown extra columns are ignored.
func ParseBanffCrossingCounts(r io.Reader) ([]BanffCrossingCountRecord, error) {
	if r == nil {
		return nil, errors.New("nil Banff CSV reader")
	}
	raw, err := io.ReadAll(io.LimitReader(r, banffMaxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read Banff CSV: %w", err)
	}
	if len(raw) > banffMaxBytes {
		return nil, fmt.Errorf("Banff CSV exceeds %d bytes", banffMaxBytes)
	}
	text, err := banffDecode(raw)
	if err != nil {
		return nil, err
	}

	cr := csv.NewReader(bytes.NewReader(text))
	cr.TrimLeadingSpace = true
	cr.FieldsPerRecord = -1

	header, err := cr.Read()
	if errors.Is(err, io.EOF) {
		return nil, errors.New("empty Banff CSV")
	}
	if err != nil {
		return nil, fmt.Errorf("read Banff header: %w", err)
	}
	cols, err := banffColumnIndex(header)
	if err != nil {
		return nil, err
	}

	var records []BanffCrossingCountRecord
	firstDataRow := true
	for {
		row, err := cr.Read()
		if errors.Is(err, io.EOF) {
			return records, nil
		}
		if err != nil {
			return nil, fmt.Errorf("read Banff CSV: %w", err)
		}
		line, _ := cr.FieldPos(0)
		if banffRowBlank(row) {
			continue
		}
		isFirst := firstDataRow
		firstDataRow = false
		if isFirst && banffIsLabelRow(row, cols) {
			continue
		}
		rec, err := banffRecord(row, cols)
		if err != nil {
			return nil, fmt.Errorf("Banff row %d: %w", line, err)
		}
		records = append(records, rec)
	}
}

// banffDecode returns UTF-8 text. Valid UTF-8 passes through unchanged so an
// already-converted file is not double-decoded; anything else is Windows-1252.
func banffDecode(raw []byte) ([]byte, error) {
	raw = bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf"))
	if utf8.Valid(raw) {
		return raw, nil
	}
	out, err := charmap.Windows1252.NewDecoder().Bytes(raw)
	if err != nil {
		return nil, fmt.Errorf("decode Banff CSV as Windows-1252: %w", err)
	}
	return out, nil
}

func banffNormalizeHeader(h string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimPrefix(h, "\uFEFF"))), " ")
}

// banffColumnIndex maps each required column name to its position and rejects a
// header that lacks any of them. Extra columns are left unmapped.
func banffColumnIndex(header []string) (map[string]int, error) {
	seen := map[string]int{}
	for i, h := range header {
		name := banffNormalizeHeader(h)
		if _, dup := seen[name]; dup {
			return nil, fmt.Errorf("duplicate Banff column %q", name)
		}
		seen[name] = i
	}
	cols := make(map[string]int, len(banffColumns))
	var missing []string
	for _, name := range banffColumns {
		idx, ok := seen[name]
		if !ok {
			missing = append(missing, name)
			continue
		}
		cols[name] = idx
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required Banff columns: %s", strings.Join(missing, ", "))
	}
	return cols, nil
}

func banffRowBlank(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func banffCell(row []string, cols map[string]int, name string) string {
	idx := cols[name]
	if idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

// banffIsLabelRow reports whether the row is the French translation of the
// header. It must be the first non-blank row after the header, and its year
// cell must be non-numeric, its first date must not parse, and no count cell
// may be numeric, so a corrupt data row is reported instead of skipped.
func banffIsLabelRow(row []string, cols map[string]int) bool {
	if _, err := strconv.Atoi(banffCell(row, cols, "year images were taken")); err == nil {
		return false
	}
	if _, err := parseBanffDate(banffCell(row, cols, "date of first image")); err == nil {
		return false
	}
	for _, name := range banffColumns[5:] {
		if _, err := strconv.Atoi(banffCell(row, cols, name)); err == nil {
			return false
		}
	}
	return true
}

// parseBanffDate parses a day/month/year date. A blank or malformed date is an
// error: a zero time.Time would read as year 1 rather than as missing.
func parseBanffDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("missing day/month/year date")
	}
	t, err := time.Parse(banffDateLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("unrecognized day/month/year date %q", s)
	}
	return t.UTC(), nil
}

func parseBanffCount(s string) (BanffCount, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return BanffCount{}, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return BanffCount{}, fmt.Errorf("invalid count %q", s)
	}
	return BanffCount{Value: n, Known: true}, nil
}

func banffRecord(row []string, cols map[string]int) (BanffCrossingCountRecord, error) {
	get := func(name string) string { return banffCell(row, cols, name) }

	year, err := strconv.Atoi(get("year images were taken"))
	if err != nil {
		return BanffCrossingCountRecord{}, fmt.Errorf("invalid image year %q", get("year images were taken"))
	}
	first, err := parseBanffDate(get("date of first image"))
	if err != nil {
		return BanffCrossingCountRecord{}, fmt.Errorf("first image: %w", err)
	}
	last, err := parseBanffDate(get("date of last image"))
	if err != nil {
		return BanffCrossingCountRecord{}, fmt.Errorf("last image: %w", err)
	}

	rec := BanffCrossingCountRecord{
		Region:     get("region"),
		Location:   get("location"),
		ImageYear:  year,
		FirstImage: first,
		LastImage:  last,
	}
	counts := []struct {
		col string
		dst *BanffCount
	}{
		{"number of hikers and runners", &rec.HikersRunners},
		{"number of horse riders", &rec.HorseRiders},
		{"number of bicyclists", &rec.Bicyclists},
		{"number of cougars", &rec.Cougars},
		{"number of elk", &rec.Elk},
		{"number of grizzly bears", &rec.GrizzlyBears},
		{"number of lynx", &rec.Lynx},
		{"number of wolves", &rec.Wolves},
		{"number of deer", &rec.Deer},
	}
	for _, c := range counts {
		v, err := parseBanffCount(get(c.col))
		if err != nil {
			return BanffCrossingCountRecord{}, fmt.Errorf("%s: %w", c.col, err)
		}
		*c.dst = v
	}
	return rec, nil
}
