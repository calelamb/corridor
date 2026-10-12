package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Broadcaster distributes real-time field reports to connected subscribers.
type Broadcaster interface {
	Subscribe() (ch <-chan FieldReport, unsubscribe func())
	Broadcast(report FieldReport)
}

type memoryBroadcaster struct {
	mu          sync.RWMutex
	subscribers map[chan FieldReport]struct{}
}

// NewBroadcaster creates a new in-memory Broadcaster.
func NewBroadcaster() Broadcaster {
	return &memoryBroadcaster{
		subscribers: make(map[chan FieldReport]struct{}),
	}
}

func (mb *memoryBroadcaster) Subscribe() (<-chan FieldReport, func()) {
	ch := make(chan FieldReport, 64)
	mb.mu.Lock()
	mb.subscribers[ch] = struct{}{}
	mb.mu.Unlock()

	unsubscribe := func() {
		mb.mu.Lock()
		delete(mb.subscribers, ch)
		mb.mu.Unlock()
		close(ch)
	}

	return ch, unsubscribe
}

func (mb *memoryBroadcaster) Broadcast(report FieldReport) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	for ch := range mb.subscribers {
		select {
		case ch <- report:
		default:
			// Non-blocking drop for slow subscriber to avoid latency spikes
		}
	}
}

var defaultBroadcaster = NewBroadcaster()

// StreamReports200Response implements StreamReportsResponseObject for text/event-stream.
type StreamReports200Response struct {
	Broadcaster Broadcaster
	Context     context.Context
}

// VisitStreamReportsResponse establishes the SSE connection and streams events.
func (r StreamReports200Response) VisitStreamReportsResponse(w http.ResponseWriter) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "Streaming unsupported by response writer")
		return nil
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	// Initial connect handshake comment
	_, _ = fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	sub, unsub := r.Broadcaster.Subscribe()
	defer unsub()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context.Done():
			return nil
		case <-ticker.C:
			_, err := fmt.Fprintf(w, ": ping\n\n")
			if err != nil {
				return nil
			}
			flusher.Flush()
		case report, ok := <-sub:
			if !ok {
				return nil
			}
			payload, err := json.Marshal(report)
			if err != nil {
				continue
			}
			_, err = fmt.Fprintf(w, "event: report\ndata: %s\n\n", payload)
			if err != nil {
				return nil
			}
			flusher.Flush()
		}
	}
}

func (s service) broadcaster() Broadcaster {
	if s.deps.Broadcaster != nil {
		return s.deps.Broadcaster
	}
	return defaultBroadcaster
}

func (s service) StreamReports(ctx context.Context, _ StreamReportsRequestObject) (StreamReportsResponseObject, error) {
	b := s.broadcaster()
	if b == nil {
		s.logFailure(ctx, "broadcaster_nil")
		return StreamReports503JSONResponse{UnavailableJSONResponse(ErrorEnvelope{
			Status: Error,
			Error:  "Broadcaster service unavailable",
		})}, nil
	}

	return StreamReports200Response{
		Broadcaster: b,
		Context:     ctx,
	}, nil
}
