package api

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/andybalholm/brotli"

	"corridor/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"
)

type Store interface {
	EventCount(context.Context) (int64, error)
	Sources(context.Context, int32, int32) ([]db.SourceSummary, int64, error)
}
type Dependencies struct {
	Store   Store
	Ready   func(context.Context) error
	Logger  *slog.Logger
	Limiter *rate.Limiter
	Timeout time.Duration
	Web     http.Handler
	Explore http.Handler
	Predict http.Handler
}
type service struct{ deps Dependencies }

func NewRouter(deps Dependencies) http.Handler {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	limiter := deps.Limiter
	if limiter == nil {
		limiter = rate.NewLimiter(100, 200)
	}
	timeout := deps.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	r := chi.NewRouter()
	r.Use(protect(logger, limiter, timeout))
	compressor := middleware.NewCompressor(6)
	compressor.SetEncoder("br", func(w io.Writer, _ int) io.Writer { return brotli.NewWriterLevel(w, 6) })
	r.Use(negotiatedCompression)
	r.Use(compressor.Handler)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if deps.Web != nil && r.URL.Path != "/v1" && !isAPIPath(r.URL.Path) {
			deps.Web.ServeHTTP(w, r)
			return
		}
		writeError(w, 404, "Not found")
	})
	if deps.Predict != nil {
		r.Get("/v1/predictions", deps.Predict.ServeHTTP)
	}
	if deps.Explore != nil {
		r.Get("/v1/explore", deps.Explore.ServeHTTP)
		r.Get("/v1/explore/config", deps.Explore.ServeHTTP)
		r.Get("/v1/migration", deps.Explore.ServeHTTP)
		r.Get("/tiles/evidence/{z}/{x}/{y}", deps.Explore.ServeHTTP)
		r.Get("/maps/region.pmtiles", deps.Explore.ServeHTTP)
	}
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) { writeError(w, 405, "Method not allowed") })
	bad := func(w http.ResponseWriter, _ *http.Request, _ error) { writeError(w, 400, "Invalid request") }
	strict := NewStrictHandlerWithOptions(service{deps}, nil, StrictHTTPServerOptions{RequestErrorHandlerFunc: bad, ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
		logger.ErrorContext(r.Context(), "response failed")
		writeError(w, 500, "Request failed")
	}})
	return HandlerWithOptions(strict, ChiServerOptions{BaseRouter: r, ErrorHandlerFunc: bad})
}
