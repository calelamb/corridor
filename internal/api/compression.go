package api

import (
	"mime"
	"net/http"
	"strconv"
	"strings"
)

// Normalize offers because chi's compressor matches tokens without interpreting q-values.
func negotiatedCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accepted := ""
		best := 0.0
		for _, encoding := range []string{"br", "gzip"} {
			if quality := encodingQuality(r.Header.Get("Accept-Encoding"), encoding); quality > best {
				accepted = encoding
				best = quality
			}
		}
		request := r.Clone(r.Context())
		request.Header.Set("Accept-Encoding", accepted)
		next.ServeHTTP(w, request)
	})
}
func encodingQuality(raw, encoding string) float64 {
	wildcard := 0.0
	for _, offer := range strings.Split(raw, ",") {
		token, params, err := mime.ParseMediaType(strings.TrimSpace(offer))
		if err != nil {
			continue
		}
		quality := 1.0
		if value, ok := params["q"]; ok {
			parsed, err := strconv.ParseFloat(value, 64)
			if err != nil || !(parsed >= 0 && parsed <= 1) {
				continue
			}
			quality = parsed
		}
		if token == encoding {
			return quality
		}
		if token == "*" {
			wildcard = quality
		}
	}
	return wildcard
}
