package cli

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

func health(ctx context.Context) error { return healthAt(ctx, "http://127.0.0.1:8080/readyz") }
func healthAt(ctx context.Context, endpoint string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.New("service unavailable")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("resource cleanup failed", "error", err)
		}
	}()
	if resp.StatusCode != 200 {
		return errors.New("dependencies unavailable")
	}
	return nil
}
