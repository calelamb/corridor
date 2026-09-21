package storage

import (
	"context"
	"corridor/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheck(t *testing.T) {
	for _, status := range []int{200, 403, 404, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }))
			defer srv.Close()
			c, err := New(config.Config{S3Endpoint: srv.URL, S3AccessKey: "test", S3SecretKey: "test-secret", S3Bucket: "test-bucket"})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			err = c.Check(ctx)
			if (err == nil) != (status == 200) {
				t.Fatalf("status %d err %v", status, err)
			}
		})
	}
}
func TestCanceled(t *testing.T) {
	c, err := New(config.Config{S3Endpoint: "http://127.0.0.1:1", S3AccessKey: "test", S3SecretKey: "test-secret", S3Bucket: "test-bucket"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if c.Check(ctx) == nil {
		t.Fatal("cancellation ignored")
	}
}
func TestInvalid(t *testing.T) {
	for _, url := range []string{"broken", "https://u:p@host", "ftp://host"} {
		if _, err := New(config.Config{S3Endpoint: url}); err == nil {
			t.Fatal("invalid accepted")
		}
	}
}

func TestSetupBoundaries(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if Setup(ctx, config.Config{}, "", "") == nil {
		t.Fatal("missing setup credentials")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusForbidden) }))
	defer srv.Close()
	cfg := config.Config{S3Endpoint: srv.URL, S3Bucket: "test-bucket", S3AccessKey: "app", S3SecretKey: "test-secret"}
	if Setup(ctx, cfg, "root", "test-secret") == nil {
		t.Fatal("denied administrator accepted")
	}
	missing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(403)
		}
	}))
	defer missing.Close()
	if Setup(ctx, config.Config{S3Endpoint: missing.URL, S3Bucket: "test-bucket", S3AccessKey: "app", S3SecretKey: "test-secret"}, "root", "test-secret") == nil {
		t.Fatal("denied create accepted")
	}
}
