package server

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestCanceledRun(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan error, 1)
	go func() { done <- Run(ctx, ln, http.NotFoundHandler(), nil) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("shutdown hung")
	}
}
func TestServeFailure(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	if err := ln.Close(); err != nil {
		t.Errorf("cleanup failed: %v", err)
	}
	if Run(context.Background(), ln, http.NotFoundHandler(), nil) == nil {
		t.Fatal("closed listener accepted")
	}
}
