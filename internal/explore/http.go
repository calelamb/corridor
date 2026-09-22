package explore

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Handler struct {
	Store   Repository
	Logger  *slog.Logger
	Basemap string
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.URL.Path == "/maps/region.pmtiles" {
		h.basemap(w, r)
		return
	}
	if r.URL.Path == "/v1/explore/config" {
		respond(w, 200, map[string]any{"status": "success", "data": map[string]any{"bounds": []float64{-117, 42, -112, 44}, "policy": Policy, "precision": "H3 resolution 6 · roughly 36 km² cells", "meaning": "Historical reports, not collision probability. Road names are source-reported; observations are not snapped.", "basemap": h.Basemap != ""}})
		return
	}
	if r.URL.Path == "/v1/migration" {
		store, ok := h.Store.(interface {
			Migration(context.Context) (json.RawMessage, error)
		})
		if !ok {
			respond(w, 503, map[string]string{"status": "error", "error": "Migration evidence unavailable"})
			return
		}
		data, err := store.Migration(r.Context())
		if err != nil {
			h.failure(w, r, err)
			return
		}
		respond(w, 200, data)
		return
	}
	values, err := parseQuery(r)
	if err != nil {
		respond(w, 400, map[string]string{"status": "error", "error": "Invalid filters"})
		return
	}
	if h.Store == nil {
		respond(w, 503, map[string]string{"status": "error", "error": "Evidence unavailable"})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/tiles/evidence/") {
		h.tile(w, r, values)
		return
	}
	data, err := h.Store.Read(r.Context(), values)
	if err != nil {
		h.failure(w, r, err)
		return
	}
	respond(w, 200, struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}{"success", data})
}
func parseQuery(r *http.Request) (Filters, error) {
	v, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return Filters{}, err
	}
	return Parse(v)
}
func (h Handler) tile(w http.ResponseWriter, r *http.Request, f Filters) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/tiles/evidence/"), "/")
	if len(parts) != 3 {
		respond(w, 400, map[string]string{"status": "error", "error": "Invalid tile"})
		return
	}
	z, e1 := strconv.Atoi(parts[0])
	x, e2 := strconv.Atoi(parts[1])
	y, e3 := strconv.Atoi(strings.TrimSuffix(parts[2], ".mvt"))
	if e1 != nil || e2 != nil || e3 != nil || z < 0 || z > 14 || x < 0 || y < 0 || x >= 1<<z || y >= 1<<z {
		respond(w, 400, map[string]string{"status": "error", "error": "Invalid tile"})
		return
	}
	data, err := h.Store.Tile(r.Context(), f, z, x, y)
	if err != nil {
		h.failure(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.mapbox-vector-tile")
	w.WriteHeader(200)
	// #nosec G705 -- ST_AsMVT binary protobuf, served as application/vnd.mapbox-vector-tile with nosniff middleware; never HTML.
	_, _ = w.Write(data)
}
func (h Handler) failure(w http.ResponseWriter, r *http.Request, err error) {
	logger := h.Logger
	if logger == nil {
		logger = slog.Default()
	}
	logger.ErrorContext(r.Context(), "exploration query failed", "error", err)
	respond(w, 503, map[string]string{"status": "error", "error": "Evidence unavailable; try again"})
}
func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func (h Handler) basemap(w http.ResponseWriter, r *http.Request) {
	if h.Basemap == "" {
		http.Error(w, "Regional basemap unavailable", http.StatusServiceUnavailable)
		return
	}
	// #nosec G304 -- operator-configured path, never derived from HTTP input.
	f, err := os.Open(h.Basemap)
	if err != nil {
		http.Error(w, "Regional basemap unavailable", http.StatusServiceUnavailable)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			slog.Warn("resource cleanup failed", "error", err)
		}
	}()
	stat, err := f.Stat()
	if err != nil {
		http.Error(w, "Regional basemap unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, "region.pmtiles", stat.ModTime(), f)
}
