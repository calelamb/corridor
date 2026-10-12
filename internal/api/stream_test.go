package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStreamReports(t *testing.T) {
	broadcaster := NewBroadcaster()
	deps := Dependencies{
		Broadcaster: broadcaster,
		ReportStore: newMockReportStore(),
	}
	router := NewRouter(deps)

	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("Receives connected header and broadcasted report within 200ms", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/v1/stream/reports", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to connect to stream: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
			t.Fatalf("expected text/event-stream, got %s", ct)
		}

		reader := bufio.NewReader(resp.Body)

		// First line should be connection handshake
		firstLine, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("failed to read handshake: %v", err)
		}
		if !strings.HasPrefix(strings.TrimSpace(firstLine), ": connected") {
			t.Fatalf("expected ': connected', got %q", firstLine)
		}

		// Channel to collect received event
		eventReceived := make(chan string, 1)

		go func() {
			var eventType, eventData string
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "event:") {
					eventType = strings.TrimSpace(strings.TrimPrefix(trimmed, "event:"))
				} else if strings.HasPrefix(trimmed, "data:") {
					eventData = strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
				} else if trimmed == "" && eventType != "" && eventData != "" {
					eventReceived <- eventData
					return
				}
			}
		}()

		// Broadcast test report
		reportID := uuid.New()
		sampleReport := FieldReport{
			Id:                reportID,
			SpeciesCommonName: "Elk",
			Latitude:          44.428,
			Longitude:         -110.588,
			ObservedAt:        time.Now().UTC(),
			Confidence:        "confirmed",
		}

		start := time.Now()
		broadcaster.Broadcast(sampleReport)

		select {
		case data := <-eventReceived:
			duration := time.Since(start)
			if duration > 200*time.Millisecond {
				t.Fatalf("stream latency exceeded 200ms: took %v", duration)
			}
			var received FieldReport
			if err := json.Unmarshal([]byte(data), &received); err != nil {
				t.Fatalf("failed to unmarshal event data: %v", err)
			}
			if received.Id != reportID || received.SpeciesCommonName != "Elk" {
				t.Fatalf("unexpected report data: %+v", received)
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("timed out waiting for broadcast event")
		}
	})

	t.Run("POST /v1/reports triggers SSE broadcast", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		streamReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/v1/stream/reports", nil)
		streamResp, err := http.DefaultClient.Do(streamReq)
		if err != nil {
			t.Fatalf("failed to connect: %v", err)
		}
		defer streamResp.Body.Close()

		reader := bufio.NewReader(streamResp.Body)
		// Consume handshake
		_, _ = reader.ReadString('\n')

		receivedCh := make(chan string, 1)
		go func() {
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "data:") {
					receivedCh <- strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
					return
				}
			}
		}()

		// Post a report
		newID := uuid.New()
		postBody, _ := json.Marshal(map[string]any{
			"id":                  newID.String(),
			"species_common_name": "Moose",
			"latitude":            43.79,
			"longitude":           -110.68,
			"observed_at":         time.Now().UTC().Format(time.RFC3339),
		})

		postReq, _ := http.NewRequest(http.MethodPost, server.URL+"/v1/reports", bytes.NewReader(postBody))
		postReq.Header.Set("Content-Type", "application/json")
		postResp, err := http.DefaultClient.Do(postReq)
		if err != nil || postResp.StatusCode != http.StatusCreated {
			t.Fatalf("failed to post report: %v, status: %d", err, postResp.StatusCode)
		}

		select {
		case data := <-receivedCh:
			var report FieldReport
			if err := json.Unmarshal([]byte(data), &report); err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			if report.Id != newID || report.SpeciesCommonName != "Moose" {
				t.Fatalf("unexpected broadcast report: %+v", report)
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("timed out waiting for POST broadcast")
		}
	})
}
