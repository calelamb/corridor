package web

import (
	"bytes"
	"html"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
)

var policyPattern = regexp.MustCompile(`<meta http-equiv="content-security-policy" content="([^"]+)"`)

const defaultPolicy = "default-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'"

func Handler(assets fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secureHeaders(w, r)
		if r.Method != "GET" && r.Method != "HEAD" {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method not allowed", 405)
			return
		}
		name, ok := assetName(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}
		data, err := fs.ReadFile(assets, name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		cache := "public, max-age=300"
		if strings.HasSuffix(name, ".html") {
			cache = "no-cache"
			if match := policyPattern.FindSubmatch(data); len(match) == 2 {
				w.Header().Set("Content-Security-Policy", html.UnescapeString(string(match[1]))+"; frame-ancestors 'none'")
			}
		} else if strings.HasPrefix(name, "_app/immutable/") {
			cache = "public, max-age=31536000, immutable"
		}
		w.Header().Set("Cache-Control", cache)
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
	})
}
func assetName(urlPath string) (string, bool) {
	if strings.ContainsAny(urlPath, "\\\x00") || strings.Contains(urlPath, "%") || strings.Contains(urlPath, "//") {
		return "", false
	}
	name := strings.TrimPrefix(urlPath, "/")
	if strings.HasPrefix(name, "v1/") || name == "v1" || strings.HasPrefix(name, "tiles/") || name == "tiles" {
		return "", false
	}
	for _, part := range strings.Split(name, "/") {
		if part == "." || part == ".." || strings.HasPrefix(part, ".") {
			return "", false
		}
	}
	if name == "" || strings.HasSuffix(name, "/") {
		name += "index.html"
	}
	return name, fs.ValidPath(name) && path.Clean(name) == name
}
func secureHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	w.Header().Set("Content-Security-Policy", defaultPolicy)
	w.Header().Set("Cache-Control", "no-store")
	if r.TLS != nil {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
	}
}
