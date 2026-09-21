package explore

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUnavailableIsNotEmpty(t *testing.T) {
	h := Handler{}
	for _, q := range []string{"", "?season=nope"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/v1/explore"+q, nil))
		expected := 503
		if q != "" {
			expected = 400
		}
		if w.Code != expected {
			t.Fatalf("status %d", w.Code)
		}
	}
}

func TestBasemapRanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.pmtiles")
	if err := os.WriteFile(path, []byte("synthetic-pmtiles-fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	h := Handler{Basemap: path}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/maps/region.pmtiles", nil)
	r.Header.Set("Range", "bytes=0-8")
	h.ServeHTTP(w, r)
	if w.Code != 206 || w.Body.String() != "synthetic" {
		t.Fatalf("range %d %q", w.Code, w.Body.String())
	}
}

type fakeRepository struct{ fail bool }

func (f fakeRepository) Read(context.Context, Filters) (json.RawMessage, error) {
	if f.fail {
		return nil, errors.New("synthetic outage")
	}
	return json.RawMessage(`{"type":"FeatureCollection","features":[]}`), nil
}
func (f fakeRepository) Tile(context.Context, Filters, int, int, int) ([]byte, error) {
	if f.fail {
		return nil, errors.New("synthetic outage")
	}
	return []byte{1, 2, 3}, nil
}
func (f fakeRepository) Migration(context.Context) (json.RawMessage, error) {
	if f.fail {
		return nil, errors.New("synthetic outage")
	}
	return json.RawMessage(`{"type":"FeatureCollection","features":[]}`), nil
}
func TestReadAndTileHTTPBoundaries(t *testing.T) {
	for _, tc := range []struct {
		path string
		fail bool
		code int
	}{{"/v1/explore", false, 200}, {"/v1/explore", true, 503}, {"/v1/explore?bad=%zz", false, 400}, {"/v1/explore/config", false, 200}, {"/v1/migration", false, 200}, {"/v1/migration", true, 503}, {"/tiles/evidence/8/46/94.mvt", false, 200}, {"/tiles/evidence/8/46/94.mvt", true, 503}, {"/tiles/evidence/99/1/1.mvt", false, 400}, {"/tiles/evidence/8/900/1.mvt", false, 400}, {"/tiles/evidence/nope", false, 400}, {"/maps/region.pmtiles", false, 503}} {
		w := httptest.NewRecorder()
		Handler{Store: fakeRepository{tc.fail}}.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		if w.Code != tc.code {
			t.Errorf("%s got %d", tc.path, w.Code)
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("stale public data cache possible")
		}
	}
}
