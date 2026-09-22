// Package ingest normalizes licensed source artifacts without publishing raw records.
package ingest

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

type Record struct {
	ID, Species, Road, Survey, Reference, Precision string
	Longitude, Latitude, Uncertainty                float64
	Start, End                                      time.Time
	Quantity                                        int64
}
type QA struct {
	Read, Accepted, Rejected, Duplicates, OutsideScope, Animals, Unsnapped int64
	Reasons                                                                map[string]int64
}
type Dataset struct {
	Records []Record
	QA      QA
}

var columns = []string{"occurrenceID", "countryCode", "scientificName", "decimalLongitude", "decimalLatitude", "coordinateUncertaintyInMeters", "numberOfRoadkill", "year", "month", "day", "roadID", "surveyType", "associatedReferences"}

func ParseRoadkill(input io.Reader) (Dataset, error) {
	reader := csv.NewReader(io.LimitReader(input, 256<<20))
	header, err := reader.Read()
	if err != nil {
		return Dataset{}, fmt.Errorf("read CSV header: %w", err)
	}
	fields := map[string]int{}
	for i, name := range header {
		if _, ok := fields[name]; ok {
			return Dataset{}, errors.New("duplicate CSV column")
		}
		fields[name] = i
	}
	for _, name := range columns {
		if _, ok := fields[name]; !ok {
			return Dataset{}, fmt.Errorf("missing column %s", name)
		}
	}
	result := Dataset{Records: []Record{}, QA: QA{Reasons: map[string]int64{}}}
	seen := map[string]bool{}
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Dataset{}, fmt.Errorf("read CSV row %d: %w", result.QA.Read+1, err)
		}
		result.QA.Read++
		if result.QA.Read > 500000 {
			return Dataset{}, errors.New("source row limit exceeded")
		}
		get := func(key string) string { return strings.TrimSpace(row[fields[key]]) }
		if get("countryCode") != "US" {
			result.QA.OutsideScope++
			continue
		}
		record, err := parseRecord(get)
		if datum, ok := fields["geodeticDatum"]; ok && row[datum] != "WGS84" && row[datum] != "WGS 84" {
			err = errors.New("unsupported datum")
		}
		if err != nil {
			result.QA.Rejected++
			result.QA.Reasons[err.Error()]++
			continue
		}
		if seen[record.ID] {
			result.QA.Duplicates++
			continue
		}
		seen[record.ID] = true
		result.Records = append(result.Records, record)
		result.QA.Accepted++
		result.QA.Animals += record.Quantity
		result.QA.Unsnapped++
	}
	return result, nil
}
func parseRecord(get func(string) string) (Record, error) {
	r := Record{ID: get("occurrenceID"), Species: get("scientificName"), Road: get("roadID"), Survey: get("surveyType"), Reference: get("associatedReferences")}
	if r.ID == "" || r.Species == "" || len(r.ID) > 200 || len(r.Species) > 200 || len(r.Road) > 300 || len(r.Reference) > 10000 {
		return Record{}, errors.New("invalid identifiers")
	}
	var err error
	r.Longitude, err = number(get("decimalLongitude"), -180, 180)
	if err != nil {
		return Record{}, errors.New("invalid longitude")
	}
	r.Latitude, err = number(get("decimalLatitude"), -90, 90)
	if err != nil {
		return Record{}, errors.New("invalid latitude")
	}
	r.Uncertainty, err = number(get("coordinateUncertaintyInMeters"), 0, 10000000)
	if err != nil {
		return Record{}, errors.New("invalid uncertainty")
	}
	n, err := number(get("numberOfRoadkill"), 1, 1000000)
	if err != nil || n != math.Trunc(n) {
		return Record{}, errors.New("invalid quantity")
	}
	r.Quantity = int64(n)
	r.Start, r.End, r.Precision, err = interval(get("year"), get("month"), get("day"))
	return r, err
}
func number(s string, min, max float64) (float64, error) {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || n < min || n > max {
		return 0, errors.New("invalid number")
	}
	return n, nil
}
func interval(ys, ms, ds string) (time.Time, time.Time, string, error) {
	fail := func() (time.Time, time.Time, string, error) {
		return time.Time{}, time.Time{}, "", errors.New("invalid date")
	}
	y, err := strconv.Atoi(ys)
	if err != nil || y < 1900 || y > time.Now().Year() {
		return fail()
	}
	m, d := 1, 1
	precision := "year"
	if ms != "" {
		m, err = strconv.Atoi(ms)
		if err != nil || m < 1 || m > 12 {
			return fail()
		}
		precision = "month"
	}
	if ds != "" {
		d, err = strconv.Atoi(ds)
		if err != nil || ms == "" || d < 1 || d > 31 {
			return fail()
		}
		precision = "day"
	}
	start := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	if start.Day() != d || int(start.Month()) != m {
		return fail()
	}
	end := start.AddDate(0, 0, 1)
	switch precision {
	case "year":
		end = start.AddDate(1, 0, 0)
	case "month":
		end = start.AddDate(0, 1, 0)
	}
	return start, end, precision, nil
}
