package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Run owns the listener and drains requests when the parent context is canceled.
func Run(ctx context.Context, listener net.Listener, handler http.Handler, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	result := make(chan error, 1)
	go func() { result <- srv.Serve(listener) }()
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdown); err != nil {
			_ = srv.Close()
			return fmt.Errorf("drain HTTP: %w", err)
		}
		err := <-result
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("stop HTTP: %w", err)
		}
		logger.Info("HTTP server stopped")
		return nil
	}
}
