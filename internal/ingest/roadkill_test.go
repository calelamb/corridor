package ingest

import (
	"strings"
	"testing"
	"time"
)

const header = "occurrenceID,countryCode,scientificName,decimalLongitude,decimalLatitude,coordinateUncertaintyInMeters,numberOfRoadkill,year,month,day,roadID,surveyType,associatedReferences\n"

func TestRoadkillIntervalsAndMultiplicity(t *testing.T) {
	d, err := ParseRoadkill(strings.NewReader(header + "one,US,Tyto alba,-115,43,30,1,2014,2,28,Interstate 84,Systematic,https://example.org/study\ntwo,US,Odocoileus hemionus,-115,43,30,3,2014,,,Interstate 84,Opportunistic,\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Records) != 2 {
		t.Fatalf("records %d", len(d.Records))
	}
	a, b := d.Records[0], d.Records[1]
	if a.Precision != "day" || a.End.Sub(a.Start) != 24*time.Hour || b.Precision != "year" || b.End.Format("2006-01-02") != "2015-01-01" || b.Quantity != 3 {
		t.Fatalf("incorrect interval/multiplicity")
	}
	if d.QA.Accepted != 2 || d.QA.Animals != 4 {
		t.Fatalf("QA %+v", d.QA)
	}
}
func TestRoadkillRejectsMalformedRows(t *testing.T) {
	rows := []string{
		"bad,US,Tyto alba,-115,43,30,1,2014,2,30,I84,Systematic,",
		"bad,US,Tyto alba,NaN,43,30,1,2014,2,1,I84,Systematic,",
		"bad,US,Tyto alba,-115,143,30,1,2014,2,1,I84,Systematic,",
		"bad,US,Tyto alba,-115,43,30,0,2014,2,1,I84,Systematic,",
		"bad,US,Tyto alba,-115,43,30,1,2014,,1,I84,Systematic,",
		"bad,US,Tyto alba,-115,43,,1,2014,2,1,I84,Systematic,",
	}
	for _, row := range rows {
		d, err := ParseRoadkill(strings.NewReader(header + row + "\n"))
		if err != nil || d.QA.Rejected != 1 || len(d.Records) != 0 {
			t.Errorf("invalid row accepted or aborted: %+v %v", d.QA, err)
		}
	}
}
func TestRoadkillDuplicateAndCountry(t *testing.T) {
	row := "one,US,Tyto alba,-115,43,30,1,2014,2,,I84,Systematic,\n"
	d, err := ParseRoadkill(strings.NewReader(header + row + row + "else,CA,Tyto alba,-115,43,30,1,2014,2,,I84,Systematic,\n"))
	if err != nil || d.QA.Duplicates != 1 || d.QA.OutsideScope != 1 || len(d.Records) != 1 || d.Records[0].Precision != "month" {
		t.Fatalf("unexpected QA %+v %v", d.QA, err)
	}
}
func TestRoadkillSchema(t *testing.T) {
	if _, err := ParseRoadkill(strings.NewReader("id,lat\nx,1\n")); err == nil {
		t.Fatal("missing schema accepted")
	}
}
