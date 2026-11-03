package ingest

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Synthetic fixture. Structure names, dates, and counts are invented for the
// test and are not rows from the Parks Canada file. Byte 0xE9 is Windows-1252
// "é"; it is not valid UTF-8 on its own.
const banffHeader = "Region,Location,Year images were taken,Date of first image,Date of last image," +
	"number of hikers and runners,number of horse riders,number of bicyclists," +
	"number of cougars,number of elk,number of grizzly bears,number of lynx,number of wolves,number of deer\n"

const banffFrenchLabelRow = "R\xe9gion,Emplacement,Ann\xe9e des images,Date de la premi\xe8re image,Date de la derni\xe8re image," +
	"nombre de randonneurs et coureurs,nombre de cavaliers,nombre de cyclistes," +
	"nombre de couguars,nombre de wapitis,nombre de grizzlis,nombre de lynx,nombre de loups,nombre de cerfs\n"

func banffFixture() []byte {
	var b bytes.Buffer
	b.WriteString(banffHeader)
	b.WriteString(banffFrenchLabelRow)
	// Blank grizzly cell on the first row; explicit 0 on the second.
	b.WriteString("Bow Valley,TCH Test Caf\xe9 UP,2015,14/10/2014,11/07/2017,12,,3,1,40,,0,2,7\n")
	b.WriteString("Massive,TCH Synthetic OP-East,2016,01/01/2016,31/12/2016,0,0,0,0,5,0,,1,\n")
	return b.Bytes()
}

func TestBanffCrossingCountArtifact(t *testing.T) {
	a := BanffCrossingCountArtifact()
	if a.License != "Open Government Licence – Canada" {
		t.Errorf("License = %q", a.License)
	}
	if a.URL != "https://open.canada.ca/data/en/dataset/d5fd45b9-7c70-4339-a68b-e60d7cbf42e8" {
		t.Errorf("URL = %q", a.URL)
	}
	if a.SHA256 != "" || a.ObjectKey != "" {
		t.Errorf("SHA256/ObjectKey must stay empty until the CSV is acquired: %q %q", a.SHA256, a.ObjectKey)
	}
	want := time.Date(2026, 11, 3, 0, 39, 38, 0, time.UTC)
	if !a.RetrievedAt.Equal(want) {
		t.Errorf("RetrievedAt = %s, want scout request time %s", a.RetrievedAt, want)
	}
}

func TestBanffDecodesWindows1252AndSkipsFrenchRow(t *testing.T) {
	recs, err := ParseBanffCrossingCounts(bytes.NewReader(banffFixture()))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("got %d records, want 2 (French label row must be skipped)", len(recs))
	}
	if recs[0].Location != "TCH Test Café UP" {
		t.Errorf("Location = %q, want decoded Windows-1252 é", recs[0].Location)
	}
	for _, r := range recs {
		if r.Region == "Région" || r.Location == "Emplacement" {
			t.Fatalf("French label row leaked into records: %+v", r)
		}
	}
	if recs[0].Region != "Bow Valley" || recs[0].ImageYear != 2015 {
		t.Errorf("row 1 = %+v", recs[0])
	}
}

func TestBanffParsesDayMonthYearDates(t *testing.T) {
	recs, err := ParseBanffCrossingCounts(bytes.NewReader(banffFixture()))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	first := time.Date(2014, time.October, 14, 0, 0, 0, 0, time.UTC)
	last := time.Date(2017, time.July, 11, 0, 0, 0, 0, time.UTC)
	if !recs[0].FirstImage.Equal(first) || !recs[0].LastImage.Equal(last) {
		t.Errorf("dates = %s..%s, want %s..%s", recs[0].FirstImage, recs[0].LastImage, first, last)
	}
	if recs[1].LastImage.Month() != time.December || recs[1].LastImage.Day() != 31 {
		t.Errorf("31/12/2016 parsed as %s", recs[1].LastImage)
	}
}

func TestBanffBlankCountIsMissingNotZero(t *testing.T) {
	recs, err := ParseBanffCrossingCounts(bytes.NewReader(banffFixture()))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if recs[0].GrizzlyBears.Known {
		t.Errorf("blank grizzly cell must be unknown, got %+v", recs[0].GrizzlyBears)
	}
	if recs[0].HorseRiders.Known {
		t.Errorf("blank horse-rider cell must be unknown, got %+v", recs[0].HorseRiders)
	}
	if !recs[1].GrizzlyBears.Known || recs[1].GrizzlyBears.Value != 0 {
		t.Errorf("explicit 0 grizzly must be known zero, got %+v", recs[1].GrizzlyBears)
	}
	if !recs[0].Elk.Known || recs[0].Elk.Value != 40 {
		t.Errorf("elk = %+v, want known 40", recs[0].Elk)
	}
	if recs[1].Deer.Known {
		t.Errorf("trailing blank deer cell must be unknown, got %+v", recs[1].Deer)
	}

	out, err := json.Marshal(recs[0])
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	if v, ok := m["grizzly_bears"]; !ok || v != nil {
		t.Errorf("JSON grizzly_bears = %v, want null for a blank cell", v)
	}
	if v := m["elk"]; v != float64(40) {
		t.Errorf("JSON elk = %v, want 40", v)
	}
}

func TestBanffExtraColumnsDoNotCreateGeometry(t *testing.T) {
	// A header with unknown extra columns, including coordinate-looking names.
	hdr := strings.TrimSuffix(banffHeader, "\n") + ",Latitude,Longitude,Notes\n"
	var b bytes.Buffer
	b.WriteString(hdr)
	b.WriteString(strings.TrimSuffix(banffFrenchLabelRow, "\n") + ",Latitude,Longitude,Notes\n")
	b.WriteString("Panorama,TCH Synthetic UP,2017,02/03/2017,04/05/2017,1,2,3,4,5,6,7,8,9,51.0,-115.0,ignored\n")

	recs, err := ParseBanffCrossingCounts(bytes.NewReader(b.Bytes()))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("got %d records, want 1", len(recs))
	}

	rt := reflect.TypeOf(recs[0])
	for i := 0; i < rt.NumField(); i++ {
		name := strings.ToLower(rt.Field(i).Name)
		for _, bad := range []string{"lat", "lon", "geom", "coord", "h3", "point"} {
			if strings.Contains(name, bad) {
				t.Errorf("record type must not carry geometry field %q", rt.Field(i).Name)
			}
		}
	}

	out, err := json.Marshal(recs[0])
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(out))
	for _, bad := range []string{"lat", "lon", "geom", "coord", "51.0", "-115"} {
		if strings.Contains(lower, bad) {
			t.Errorf("JSON must not contain %q: %s", bad, out)
		}
	}
}

func TestBanffAcceptsUTF8Input(t *testing.T) {
	// Already-decoded text must not be double-decoded into mojibake.
	in := banffHeader + "Région,Emplacement,Année,Date,Date,n,n,n,n,n,n,n,n,n\n" +
		"Lake Louise,TCH Café UP,2014,14/10/2014,15/10/2014,1,1,1,1,1,1,1,1,1\n"
	recs, err := ParseBanffCrossingCounts(strings.NewReader(in))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(recs) != 1 || recs[0].Location != "TCH Café UP" {
		t.Errorf("recs = %+v", recs)
	}
}

func TestBanffRejectsMissingColumns(t *testing.T) {
	in := "Region,Location,Year images were taken\nBow Valley,TCH X UP,2015\n"
	if _, err := ParseBanffCrossingCounts(strings.NewReader(in)); err == nil {
		t.Fatal("expected error for missing required columns")
	}
	if _, err := ParseBanffCrossingCounts(strings.NewReader("")); err == nil {
		t.Fatal("expected error for empty input")
	}
	if _, err := ParseBanffCrossingCounts(nil); err == nil {
		t.Fatal("expected error for nil reader")
	}
}

func TestBanffRejectsNonNumericCount(t *testing.T) {
	in := banffHeader + banffFrenchLabelRow +
		"Bow Valley,TCH X UP,2015,14/10/2014,11/07/2017,many,,3,1,40,,0,2,7\n"
	if _, err := ParseBanffCrossingCounts(bytes.NewReader([]byte(in))); err == nil {
		t.Fatal("expected error for non-numeric count cell")
	}
}

func TestBanffDateCases(t *testing.T) {
	cases := []struct {
		name    string
		first   string
		want    time.Time
		wantErr bool
	}{
		{"day first, ambiguous", "03/04/2017", time.Date(2017, time.April, 3, 0, 0, 0, 0, time.UTC), false},
		{"day above 12", "14/10/2014", time.Date(2014, time.October, 14, 0, 0, 0, 0, time.UTC), false},
		{"single digit parts", "3/4/2017", time.Date(2017, time.April, 3, 0, 0, 0, 0, time.UTC), false},
		{"month-first rejected", "10/14/2014", time.Time{}, true},
		{"ISO rejected", "2014-10-14", time.Time{}, true},
		{"two-digit year rejected", "14/10/14", time.Time{}, true},
		{"blank is not a zero date", "", time.Time{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := banffHeader + banffFrenchLabelRow +
				"Bow Valley,TCH X UP,2015," + tc.first + ",11/07/2017,1,1,1,1,1,1,1,1,1\n"
			recs, err := ParseBanffCrossingCounts(strings.NewReader(in))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", recs)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if !recs[0].FirstImage.Equal(tc.want) {
				t.Errorf("FirstImage = %s, want %s", recs[0].FirstImage, tc.want)
			}
		})
	}
}

func TestBanffCountCases(t *testing.T) {
	cases := []struct {
		name      string
		cell      string
		wantKnown bool
		wantValue int
		wantErr   bool
	}{
		{"blank", "", false, 0, false},
		{"whitespace only", "  ", false, 0, false},
		{"digit zero", "0", true, 0, false},
		{"positive", "12", true, 12, false},
		{"padded", " 7 ", true, 7, false},
		{"negative", "-1", false, 0, true},
		{"decimal", "1.5", false, 0, true},
		{"text", "n/a", false, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := banffHeader + banffFrenchLabelRow +
				"Bow Valley,TCH X UP,2015,14/10/2014,11/07/2017,\"" + tc.cell + "\",1,1,1,1,1,1,1,1\n"
			recs, err := ParseBanffCrossingCounts(strings.NewReader(in))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", recs)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			got := recs[0].HikersRunners
			if got.Known != tc.wantKnown || got.Value != tc.wantValue {
				t.Errorf("count = %+v, want Known=%v Value=%d", got, tc.wantKnown, tc.wantValue)
			}
			out, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			wantJSON := "null"
			if tc.wantKnown {
				wantJSON = strconv.Itoa(tc.wantValue)
			}
			if string(out) != wantJSON {
				t.Errorf("JSON = %s, want %s", out, wantJSON)
			}
		})
	}
}

func TestBanffLabelRowHandling(t *testing.T) {
	const dataRow = "Bow Valley,TCH X UP,2015,14/10/2014,11/07/2017,1,1,1,1,1,1,1,1,1\n"
	cases := []struct {
		name    string
		body    string
		want    int
		wantErr bool
	}{
		{"French row skipped", banffFrenchLabelRow + dataRow, 1, false},
		{"no French row keeps first data row", dataRow, 1, false},
		{"blank-cell row before label row", ",,,,,,,,,,,,,\n" + banffFrenchLabelRow + dataRow, 1, false},
		{"corrupt first data row is an error, not skipped",
			"Bow Valley,TCH X UP,bad,bad,bad,1,1,1,1,1,1,1,1,1\n" + dataRow, 0, true},
		{"French-looking row after data is an error", dataRow + banffFrenchLabelRow, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recs, err := ParseBanffCrossingCounts(strings.NewReader(banffHeader + tc.body))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", recs)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if len(recs) != tc.want {
				t.Errorf("got %d records, want %d", len(recs), tc.want)
			}
		})
	}
}

func TestBanffWindows1252Byte(t *testing.T) {
	in := banffHeader + banffFrenchLabelRow +
		"Bow Valley,Caf\xe9 \x92Underpass\x92,2015,14/10/2014,11/07/2017,1,1,1,1,1,1,1,1,1\n"
	recs, err := ParseBanffCrossingCounts(strings.NewReader(in))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if want := "Café \u2019Underpass\u2019"; recs[0].Location != want {
		t.Errorf("Location = %q, want %q", recs[0].Location, want)
	}
}

func TestBanffRejectsDuplicateColumn(t *testing.T) {
	in := strings.TrimSuffix(banffHeader, "\n") + ",Number of Elk\n" +
		"Bow Valley,TCH X UP,2015,14/10/2014,11/07/2017,1,1,1,1,1,1,1,1,1,1\n"
	if _, err := ParseBanffCrossingCounts(strings.NewReader(in)); err == nil {
		t.Fatal("expected error for duplicate required column")
	}
}

func TestBanffRejectsOversizeInput(t *testing.T) {
	in := banffHeader + strings.Repeat("x", banffMaxBytes)
	if _, err := ParseBanffCrossingCounts(strings.NewReader(in)); err == nil {
		t.Fatal("expected error for input over the size limit")
	}
}
