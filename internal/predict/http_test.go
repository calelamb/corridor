package predict

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
)

type fakeStore struct {
	rows []Sample
	err  error
}

func (f fakeStore) Samples(context.Context) ([]Sample, error) { return f.rows, f.err }
func TestBoundary(t *testing.T) {
	for _, tc := range []struct {
		query  string
		store  Repository
		status int
	}{{"", fakeStore{synthetic(), nil}, 200}, {"?study=pequop", fakeStore{synthetic(), nil}, 200}, {"?study=unknown", nil, 400}, {"?study=pequop&study=pequop", nil, 400}, {"?species=deer", nil, 400}, {"?x=%zz", nil, 400}, {"", fakeStore{nil, errors.New("private database error")}, 503}, {"", fakeStore{}, 422}, {"", nil, 503}} {
		r := httptest.NewRecorder()
		Handler{Store: tc.store}.ServeHTTP(r, httptest.NewRequest("GET", "/v1/predictions"+tc.query, nil))
		if r.Code != tc.status {
			t.Fatalf("%s: %d %s", tc.query, r.Code, r.Body.String())
		}
		if r.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("cached")
		}
	}
}
func TestResultRanking(t *testing.T) {
	rows := synthetic()
	for i := range rows {
		rows[i].Geometry = json.RawMessage(`{"type":"MultiPolygon","coordinates":[]}`)
		rows[i].Road = i%4 == 0
	}
	m, err := Fit(rows)
	if err != nil {
		t.Fatal(err)
	}
	r := Build(m)
	if r.CandidateCount != 40 {
		t.Fatal(r.CandidateCount)
	}
	for _, f := range r.Features {
		if f.Properties.Score < 0 || f.Properties.Score > 100 || f.Properties.Road != (f.Properties.Rank > 0) {
			t.Fatalf("bad rank %+v", f)
		}
	}
}
