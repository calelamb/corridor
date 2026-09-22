//go:build integration

package storage

import (
	"bytes"
	"context"
	"corridor/internal/config"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPrivateSetup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: testcontainers.ContainerRequest{Image: "corridor-minio:foundation", ExposedPorts: []string{"9000/tcp"}, Env: map[string]string{"MINIO_ROOT_USER": "synthetic-admin", "MINIO_ROOT_PASSWORD": "synthetic-admin-password"}, WaitingFor: wait.ForHTTP("/minio/health/live").WithPort("9000/tcp")}, Started: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Terminate(context.Background()); err != nil {
			t.Error(err)
		}
	})
	host, err := c.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := c.MappedPort(ctx, "9000/tcp")
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{S3Endpoint: "http://" + host + ":" + port.Port(), S3Bucket: "synthetic-private", S3AccessKey: "synthetic-app", S3SecretKey: "synthetic-app-password"}
	for range 2 {
		if err := Setup(ctx, cfg, "synthetic-admin", "synthetic-admin-password"); err != nil {
			t.Fatal(err)
		}
	}
	app, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Check(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := app.api.PutObject(ctx, cfg.S3Bucket, "forbidden", bytes.NewReader([]byte("test")), 4, minio.PutObjectOptions{}); err == nil {
		t.Fatal("HTTP app may write raw objects")
	}
	admin, err := New(config.Config{S3Endpoint: cfg.S3Endpoint, S3Bucket: cfg.S3Bucket, S3AccessKey: "synthetic-admin", S3SecretKey: "synthetic-admin-password"})
	if err != nil {
		t.Fatal(err)
	}
	key := "sha256/" + strings.Repeat("a", 64) + "/synthetic.csv"
	for range 2 {
		if err := admin.Archive(ctx, key, bytes.NewReader([]byte("test")), 4); err != nil {
			t.Fatal(err)
		}
	}
	if err := admin.Archive(ctx, "bad", bytes.NewReader(nil), 0); err == nil {
		t.Fatal("invalid archive accepted")
	}
	if err := app.Archive(ctx, "sha256/"+strings.Repeat("b", 64)+"/synthetic.csv", bytes.NewReader([]byte("test")), 4); err == nil {
		t.Fatal("public app archived raw data")
	}
	if _, err := admin.api.PutObject(ctx, cfg.S3Bucket, "test.txt", bytes.NewReader([]byte("synthetic test")), 14, minio.PutObjectOptions{}); err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", cfg.S3Endpoint+"/"+cfg.S3Bucket+"/test.txt", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	}()
	if resp.StatusCode != 403 {
		t.Fatalf("anonymous status %d", resp.StatusCode)
	}
}
