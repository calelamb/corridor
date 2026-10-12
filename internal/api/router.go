package api

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/andybalholm/brotli"

	"corridor/internal/db"
	"corridor/internal/geo"
	"corridor/internal/prioritize"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/paulmach/orb"
	"golang.org/x/time/rate"
)

type Store interface {
	EventCount(context.Context) (int64, error)
	Sources(context.Context, int32, int32) ([]db.SourceSummary, int64, error)
}

type SegmentStore interface {
	Segments(ctx context.Context, filter db.SegmentFilter) ([]db.SegmentSummary, error)
	Segment(ctx context.Context, id uuid.UUID) (*db.SegmentDetail, error)
}

type MitigationProvider interface {
	GenerateCandidates(ctx context.Context, budget float64, minBCR *float64, state *string) ([]prioritize.CandidateSite, error)
	CandidateSite(ctx context.Context, id uuid.UUID) (*prioritize.CandidateSite, error)
}

type ReportStore interface {
	CreateReport(ctx context.Context, report FieldReportRecord) (*FieldReportRecord, bool, error)
	GetReport(ctx context.Context, id uuid.UUID) (*FieldReportRecord, error)
	UpdateReportPhoto(ctx context.Context, id uuid.UUID, photoKey string, suggestions []SpeciesSuggestion) (*FieldReportRecord, error)
}

type PhotoStore interface {
	PutPhoto(ctx context.Context, key string, data []byte) error
	GetPhoto(ctx context.Context, key string) ([]byte, error)
}

type SpeciesClassifier interface {
	Classify(ctx context.Context, imgBytes []byte) ([]SpeciesSuggestion, error)
}

type Snapper func(ctx context.Context, point orb.Point, maxDistanceMeters float64) (segmentID uuid.UUID, snapDistance float64, snappedPoint orb.Point, err error)

type Deduplicator func(ctx context.Context, speciesID string, geom orb.Point, observedAt time.Time, distanceThresholdM float64, timeThresholdHours float64) (dedupGroupID *uuid.UUID, isDuplicate bool, err error)

type Dependencies struct {
	Store              Store
	SegmentStore       SegmentStore
	MitigationProvider MitigationProvider
	ReportStore        ReportStore
	PhotoStore         PhotoStore
	SpeciesClassifier  SpeciesClassifier
	Snapper            Snapper
	Deduplicator       Deduplicator
	Ready              func(context.Context) error
	Logger             *slog.Logger
	Limiter            *rate.Limiter
	Timeout            time.Duration
	Web                http.Handler
	Explore            http.Handler
	Predict            http.Handler
}
type service struct{ deps Dependencies }

func (s service) segmentStore() SegmentStore {
	if s.deps.SegmentStore != nil {
		return s.deps.SegmentStore
	}
	if ss, ok := s.deps.Store.(SegmentStore); ok {
		return ss
	}
	return nil
}

func (s service) mitigationProvider() MitigationProvider {
	if s.deps.MitigationProvider != nil {
		return s.deps.MitigationProvider
	}
	if dbs, ok := s.deps.Store.(db.Store); ok && dbs.Pool != nil {
		return prioritize.NewEngine(dbs.Pool, prioritize.DefaultConfig())
	}
	return nil
}

func (s service) reportStore() ReportStore {
	if s.deps.ReportStore != nil {
		return s.deps.ReportStore
	}
	if rs, ok := s.deps.Store.(ReportStore); ok {
		return rs
	}
	if dbs, ok := s.deps.Store.(db.Store); ok && dbs.Pool != nil {
		return dbReportStoreAdapter{dbs}
	}
	return nil
}

func (s service) photoStore() PhotoStore {
	if s.deps.PhotoStore != nil {
		return s.deps.PhotoStore
	}
	return nil
}

func (s service) speciesClassifier() SpeciesClassifier {
	if s.deps.SpeciesClassifier != nil {
		return s.deps.SpeciesClassifier
	}
	return defaultClassifier{}
}

func (s service) snapper() Snapper {
	if s.deps.Snapper != nil {
		return s.deps.Snapper
	}
	if dbs, ok := s.deps.Store.(db.Store); ok && dbs.Pool != nil {
		return func(ctx context.Context, point orb.Point, maxDistanceMeters float64) (uuid.UUID, float64, orb.Point, error) {
			return geo.SnapToSegment(ctx, dbs.Pool, point, maxDistanceMeters)
		}
	}
	return nil
}

func (s service) deduplicator() Deduplicator {
	if s.deps.Deduplicator != nil {
		return s.deps.Deduplicator
	}
	if dbs, ok := s.deps.Store.(db.Store); ok && dbs.Pool != nil {
		return func(ctx context.Context, speciesID string, geom orb.Point, observedAt time.Time, distanceThresholdM, timeThresholdHours float64) (*uuid.UUID, bool, error) {
			return geo.DeduplicateEvents(ctx, dbs.Pool, speciesID, geom, observedAt, distanceThresholdM, timeThresholdHours)
		}
	}
	return nil
}

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
