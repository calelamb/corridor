package ingest

import (
	"encoding/csv"
	"math"
	"strings"
	"testing"
)

// FuzzParseRoadkillRow tests roadkill ingestion against valid, boundary, and
// adversarially malformed CSV rows. It asserts the parser never panics and cleanly
// handles unexpected quotes, commas, null bytes, unicode, and corrupt numbers.
func FuzzParseRoadkillRow(f *testing.F) {
	// Seed with valid CSV rows
	f.Add("one,US,Tyto alba,-115,43,30,1,2014,2,28,Interstate 84,Systematic,https://example.org/study")
	f.Add("two,US,Odocoileus hemionus,-115,43,30,3,2014,,,Interstate 84,Opportunistic,")
	f.Add("\"three, with comma\",US,\"Cervus canadensis\",-110.5,44.5,15,2,2023,6,15,\"US-191\",\"Systematic\",\"https://doi.org/10.1000/182\"")
	f.Add("four,US,Canis lupus,-110.75,44.60,5,1,2022,1,,Yellowstone Loop,Systematic,")

	// Boundary and validation error rows
	f.Add("bad,US,Tyto alba,-115,43,30,1,2014,2,30,I84,Systematic,")
	f.Add("bad,US,Tyto alba,NaN,43,30,1,2014,2,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,Inf,43,30,1,2014,2,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-Inf,43,30,1,2014,2,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-115,143,30,1,2014,2,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-115,-95,30,1,2014,2,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-185,43,30,1,2014,2,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,185,43,30,1,2014,2,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-115,43,30,0,2014,2,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-115,43,30,-5,2014,2,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-115,43,30,1.5,2014,2,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-115,43,30,1,2014,,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-115,43,,1,2014,2,1,I84,Systematic,")
	f.Add("ca_row,CA,Alces alces,-115,43,30,1,2014,2,28,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-115,43,30,1,1899,2,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-115,43,30,1,2099,2,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-115,43,30,1,2014,13,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-115,43,30,1,2014,0,1,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-115,43,30,1,2014,2,0,I84,Systematic,")
	f.Add("bad,US,Tyto alba,-115,43,30,1,2014,2,32,I84,Systematic,")

	// Malformed rows: quotes, commas, null bytes, unicode, whitespace
	f.Add("")
	f.Add("   ")
	f.Add("unclosed \"quote,US,Tyto alba,-115,43,30,1,2014,2,28,I84,Systematic,")
	f.Add("null\x00byte,US,Tyto\x00alba,-115,43,30,1,2014,2,28,I84,Systematic,")
	f.Add("deer_\U0001f98c,US,Odocoileus virginianus \u00e9\u00e8,-115.5,43.2,30,1,2021,5,12,US-20,Systematic,")
	f.Add("\"escaped \"\"quote\"\",US,Tyto alba,-115,43,30,1,2014,2,28,I84,Systematic,")
	f.Add("commas,,,too,,,many,,,commas,,,extra1,extra2,extra3")
	f.Add("short,row")
	f.Add(",,,,,,,,,,,,")
	f.Add(strings.Repeat("A", 1000) + ",US,Tyto alba,-115,43,30,1,2014,2,28,I84,Systematic,")

	f.Fuzz(func(t *testing.T, row string) {
		// Test 1: Full CSV pipeline with prepended standard header
		fullCSV := header + row + "\n"
		d, err := ParseRoadkill(strings.NewReader(fullCSV))
		if err == nil {
			// Assert QA counters are structurally consistent
			if d.QA.Accepted+d.QA.Rejected+d.QA.OutsideScope != d.QA.Read {
				t.Fatalf("QA counters inconsistent: %+v", d.QA)
			}
			for _, rec := range d.Records {
				if math.IsNaN(rec.Longitude) || math.IsInf(rec.Longitude, 0) || rec.Longitude < -180 || rec.Longitude > 180 {
					t.Fatalf("invalid longitude in accepted record: %v", rec.Longitude)
				}
				if math.IsNaN(rec.Latitude) || math.IsInf(rec.Latitude, 0) || rec.Latitude < -90 || rec.Latitude > 90 {
					t.Fatalf("invalid latitude in accepted record: %v", rec.Latitude)
				}
				if math.IsNaN(rec.Uncertainty) || math.IsInf(rec.Uncertainty, 0) || rec.Uncertainty < 0 {
					t.Fatalf("invalid uncertainty in accepted record: %v", rec.Uncertainty)
				}
				if rec.Quantity < 1 {
					t.Fatalf("invalid quantity in accepted record: %d", rec.Quantity)
				}
				if rec.Start.IsZero() || rec.End.IsZero() || rec.End.Before(rec.Start) {
					t.Fatalf("invalid date interval: %v to %v", rec.Start, rec.End)
				}
				if rec.Precision != "year" && rec.Precision != "month" && rec.Precision != "day" {
					t.Fatalf("invalid date precision: %s", rec.Precision)
				}
				if rec.ID == "" || len(rec.ID) > 200 || rec.Species == "" || len(rec.Species) > 200 {
					t.Fatalf("invalid record identifiers: ID=%q, Species=%q", rec.ID, rec.Species)
				}
			}
		}

		// Test 2: Parse raw row directly using CSV reader and ParseRoadkillRow
		r := csv.NewReader(strings.NewReader(row))
		if fields, err := r.Read(); err == nil {
			rec, pErr := ParseRoadkillRow(fields)
			if pErr == nil {
				if rec.Quantity < 1 {
					t.Fatalf("ParseRoadkillRow produced invalid quantity: %d", rec.Quantity)
				}
				if math.IsNaN(rec.Longitude) || math.IsInf(rec.Longitude, 0) || rec.Longitude < -180 || rec.Longitude > 180 {
					t.Fatalf("ParseRoadkillRow produced invalid longitude: %v", rec.Longitude)
				}
				if math.IsNaN(rec.Latitude) || math.IsInf(rec.Latitude, 0) || rec.Latitude < -90 || rec.Latitude > 90 {
					t.Fatalf("ParseRoadkillRow produced invalid latitude: %v", rec.Latitude)
				}
			}
		}

		// Test 3: Raw input directly as an entire CSV (testing header parsing / malformed files)
		_, _ = ParseRoadkill(strings.NewReader(row))
	})
}
