package web

import (
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestStaticRouting(t *testing.T) {
	files := fstest.MapFS{
		"index.html":             {Data: []byte(`<html><meta http-equiv="content-security-policy" content="script-src 'self' 'sha256-test'"></html>`)},
		"data/index.html":        {Data: []byte("<html>Data</html>")},
		"methodology/index.html": {Data: []byte("<html>Methodology</html>")},
		"report/index.html":      {Data: []byte("<html>Report</html>")},
		"_app/immutable/test.js": {Data: []byte("export default 1")},
		"robots.txt":             {Data: []byte("User-agent: *")},
	}
	h := Handler(files)
	for _, tt := range []struct {
		path, method string
		code         int
		location     string
	}{
		{"/", "GET", 200, ""},
		{"/index.html", "GET", 200, ""},
		{"/data", "GET", 301, "/data/"},
		{"/data/", "GET", 200, ""},
		{"/methodology", "GET", 301, "/methodology/"},
		{"/methodology/", "GET", 200, ""},
		{"/report", "GET", 301, "/report/"},
		{"/report/", "GET", 200, ""},
		{"/data?tab=summary", "GET", 301, "/data/?tab=summary"},
		{"/nonexistent", "GET", 404, ""},
		{"/_app/immutable/test.js", "GET", 200, ""},
		{"/_app/missing.js", "GET", 404, ""},
		{"/", "HEAD", 200, ""},
		{"/", "POST", 405, ""},
		{"/../index.html", "GET", 404, ""},
		{"/%2e%2e/index.html", "GET", 404, ""},
		{"/_app/", "GET", 404, ""},
		{"/v1/missing", "GET", 404, ""},
		{"/tiles/events/0/0/0.mvt", "GET", 404, ""},
		{"/robots.txt", "GET", 200, ""},
	} {
		t.Run(tt.method+tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
			if w.Code != tt.code {
				t.Fatalf("got %d", w.Code)
			}
			if tt.location != "" && w.Header().Get("Location") != tt.location {
				t.Fatalf("expected Location %q, got %q", tt.location, w.Header().Get("Location"))
			}
			if tt.code != 200 && strings.Contains(w.Body.String(), "<html") {
				t.Fatal("HTML fallback")
			}
			if tt.method == "HEAD" && w.Body.Len() != 0 {
				t.Fatal("HEAD body")
			}
			if w.Header().Get("X-Frame-Options") != "DENY" {
				t.Fatal("missing security headers")
			}
			if tt.path == "/" && tt.code == 200 && w.Header().Get("Cache-Control") != "no-cache" {
				t.Fatal("HTML cached")
			}
			if strings.Contains(tt.path, "immutable/test") && !strings.Contains(w.Header().Get("Cache-Control"), "immutable") {
				t.Fatal("asset not cached")
			}
		})
	}
}
func TestEmbeddedAssets(t *testing.T) {
	assets, err := Assets()
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRecorder()
	Handler(assets).ServeHTTP(r, httptest.NewRequest("GET", "/", nil))
	if r.Code != 200 || !strings.Contains(r.Body.String(), "corridor") {
		t.Fatal("missing embedded application")
	}

	for _, p := range []string{"/methodology", "/report", "/data"} {
		// Clean URL without trailing slash redirects with 301
		rec := httptest.NewRecorder()
		Handler(assets).ServeHTTP(rec, httptest.NewRequest("GET", p, nil))
		if rec.Code != 301 || rec.Header().Get("Location") != p+"/" {
			t.Fatalf("path %s: expected 301 redirect to %s/, got %d with Location %q", p, p, rec.Code, rec.Header().Get("Location"))
		}

		// Trailing slash URL serves index.html with 200 OK
		recSlash := httptest.NewRecorder()
		Handler(assets).ServeHTTP(recSlash, httptest.NewRequest("GET", p+"/", nil))
		if recSlash.Code != 200 {
			t.Fatalf("path %s/: expected 200 OK, got %d", p, recSlash.Code)
		}
		if !strings.Contains(recSlash.Body.String(), "<html") {
			t.Fatalf("path %s/: expected HTML body", p)
		}
	}
}

func TestPrecompressedAssets(t *testing.T) {
	h := Handler(fstest.MapFS{"_app/immutable/a.js": {Data: []byte("script")}, "_app/immutable/a.js.br": {Data: []byte("synthetic encoded fixture")}})
	for _, encoding := range []string{"br", "br;q=0", ""} {
		req := httptest.NewRequest("GET", "/_app/immutable/a.js", nil)
		req.Header.Set("Accept-Encoding", encoding)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if (w.Header().Get("Content-Encoding") == "br") != (encoding == "br") {
			t.Fatalf("encoding %q: %s", encoding, w.Header().Get("Content-Encoding"))
		}
		if !strings.Contains(w.Header().Get("Content-Type"), "javascript") {
			t.Fatal("lost MIME type")
		}
	}
}
