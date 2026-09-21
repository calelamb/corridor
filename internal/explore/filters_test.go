package explore

import (
	"net/url"
	"testing"
)

func TestFiltersBoundary(t *testing.T) {
	for _, q := range []string{"bbox=NaN,0,1,2", "bbox=10,0,0,1", "bbox=-181,0,1,2", "season=wet", "start=2050", "start=2020&end=2010", "limit=501", "offset=-1", "secret=1", "species=a&species=b", "cell=private", "q=" + string(make([]byte, 301))} {
		v, err := url.ParseQuery(q)
		if err == nil {
			if _, err = Parse(v); err == nil {
				t.Errorf("accepted %q", q)
			}
		}
	}
	f, err := Parse(url.Values{"bbox": {"-118,40,-109,46"}, "season": {"winter"}, "start": {"2010"}, "end": {"2020"}})
	if err != nil || f.Start != 2010 || f.Limit != 100 {
		t.Fatalf("valid filter %+v %v", f, err)
	}
}
