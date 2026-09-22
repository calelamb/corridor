package predict

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
)

type Handler struct {
	Store  Repository
	Logger *slog.Logger
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(params) > 1 {
		respond(w, 400, nil, "Unsupported scenario")
		return
	}
	for k, v := range params {
		if k != "study" || len(v) != 1 || v[0] != "pequop" {
			respond(w, 400, nil, "Only the Pequop mule-deer study is supported")
			return
		}
	}
	if h.Store == nil {
		respond(w, 503, nil, "Movement analysis unavailable")
		return
	}
	rows, err := h.Store.Samples(r.Context())
	if err != nil {
		h.failed(w, r, err)
		return
	}
	m, err := FitContext(r.Context(), rows)
	if errors.Is(err, ErrSupport) {
		respond(w, 422, nil, "Insufficient released movement and road data")
		return
	}
	if err != nil {
		h.failed(w, r, err)
		return
	}
	if err = r.Context().Err(); err != nil {
		h.failed(w, r, err)
		return
	}
	respond(w, 200, Build(m), "")
}
func (h Handler) failed(w http.ResponseWriter, r *http.Request, err error) {
	logger := h.Logger
	if logger == nil {
		logger = slog.Default()
	}
	logger.ErrorContext(r.Context(), "movement analysis failed", "error", err)
	respond(w, 503, nil, "Movement analysis unavailable; try again")
}
func respond(w http.ResponseWriter, status int, data any, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	result := "success"
	var failure *string
	if message != "" {
		result = "error"
		failure = &message
	}
	_ = json.NewEncoder(w).Encode(struct {
		Status string  `json:"status"`
		Data   any     `json:"data"`
		Error  *string `json:"error"`
		Meta   any     `json:"meta"`
	}{result, data, failure, nil})
}
