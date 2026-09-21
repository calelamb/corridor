package web

import (
	"io/fs"
	"net/http"
	"strings"
)

// Prebuilt variants serve common browser encodings; weighted offers fall back to middleware.
func compressed(assets fs.FS, r *http.Request, name string, original []byte) ([]byte, string) {
	offers := strings.Split(r.Header.Get("Accept-Encoding"), ",")
	for _, encoding := range []struct{ name, suffix string }{{"br", ".br"}, {"gzip", ".gz"}} {
		for _, offer := range offers {
			if strings.TrimSpace(offer) != encoding.name {
				continue
			}
			if data, err := fs.ReadFile(assets, name+encoding.suffix); err == nil {
				return data, encoding.name
			}
		}
	}
	return original, ""
}
