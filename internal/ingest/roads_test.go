package ingest

import (
	"strings"
	"testing"
)

func TestParseRoads(t *testing.T) {
	valid := `{"elements":[{"type":"way","id":1,"tags":{"highway":"motorway","ref":"I 80"},"geometry":[{"lon":-115,"lat":41},{"lon":-114.9,"lat":41.1}]}]}`
	roads, err := ParseRoads(strings.NewReader(valid))
	if err != nil || len(roads) != 1 {
		t.Fatalf("%v %v", roads, err)
	}
	for _, s := range []string{`{}`, `{"remark":"timeout","elements":[]}`, strings.Replace(valid, `"lat":41`, `"lat":141`, 1), strings.Replace(valid, `"motorway"`, `"residential"`, 1), strings.Replace(valid, `"id":1`, `"id":0`, 1)} {
		if _, err = ParseRoads(strings.NewReader(s)); err == nil {
			t.Fatalf("accepted invalid source")
		}
	}
}
