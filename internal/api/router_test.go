package api

import (
	"context"
	"corridor/internal/db"
	"errors"
	"golang.org/x/time/rate"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testStore struct {
	count int64
	err   error
}

func (s testStore) EventCount(context.Context) (int64, error) { return s.count, s.err }
func (s testStore) Sources(context.Context, int32, int32) ([]db.SourceSummary, int64, error) {
	return []db.SourceSummary{}, 0, s.err
}
func request(h http.Handler, method, path string) *httptest.ResponseRecorder {
	r := httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest(method, path, nil))
	return r
}
func TestPublicAPI(t *testing.T) {
	for _, tt := range []struct {
		path, method, contains string
		status                 int
	}{{"/healthz", "GET", "success", 200}, {"/readyz", "GET", "success", 200}, {"/v1/coverage", "GET", "empty", 200}, {"/v1/sources", "GET", `"total":0`, 200}, {"/v1/sources?limit=0", "GET", "error", 400}, {"/v1/sources?limit=101", "GET", "error", 400}, {"/v1/sources?offset=-1", "GET", "error", 400}, {"/v1/sources?limit=no", "GET", "error", 400}, {"/v1/sources?offset=100001", "GET", "error", 400}, {"/v1/missing", "GET", "error", 404}, {"/healthz", "POST", "error", 405}} {
		t.Run(tt.path+tt.method, func(t *testing.T) {
			h := NewRouter(Dependencies{Store: testStore{}, Ready: func(context.Context) error { return nil }})
			r := request(h, tt.method, tt.path)
			if r.Code != tt.status || !strings.Contains(r.Body.String(), tt.contains) {
				t.Fatalf("%d %s", r.Code, r.Body.String())
			}
			if r.Header().Get("X-Request-ID") == "" {
				t.Fatal("no request id")
			}
		})
	}
	r := request(NewRouter(Dependencies{Store: testStore{count: 1}}), "GET", "/v1/coverage")
	if !strings.Contains(r.Body.String(), "unmodeled") {
		t.Fatal(r.Body)
	}
}
func TestFailuresDoNotLeak(t *testing.T) {
	for _, deps := range []Dependencies{{}, {Store: testStore{err: errors.New("private-dsn")}, Ready: func(context.Context) error { return errors.New("private-dsn") }}} {
		for _, path := range []string{"/readyz", "/v1/coverage", "/v1/sources"} {
			r := request(NewRouter(deps), "GET", path)
			if r.Code != 503 || strings.Contains(r.Body.String(), "private-dsn") {
				t.Fatalf("%d %s", r.Code, r.Body)
			}
		}
	}
}
func TestRateLimit(t *testing.T) {
	h := NewRouter(Dependencies{Limiter: rate.NewLimiter(0, 1)})
	if request(h, "GET", "/healthz").Code != 200 {
		t.Fatal("first")
	}
	r := request(h, "GET", "/healthz")
	if r.Code != 429 || r.Header().Get("Retry-After") == "" {
		t.Fatal(r)
	}
}
func TestCanceledReady(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	h := NewRouter(Dependencies{Ready: func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }})
	r := httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("GET", "/readyz", nil).WithContext(ctx))
	if r.Code != 503 {
		t.Fatal(r.Code)
	}
}
func TestPanicRecovered(t *testing.T) {
	h := NewRouter(Dependencies{Ready: func(context.Context) error { panic("secret") }})
	r := request(h, "GET", "/readyz")
	if r.Code != 500 || strings.Contains(r.Body.String(), "secret") {
		t.Fatal(r)
	}
}
