package main

import (
	"context"
	"corridor/internal/api"
	"corridor/internal/config"
	"corridor/internal/db"
	"corridor/internal/server"
	"corridor/internal/storage"
	"errors"
	"golang.org/x/time/rate"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func run(ctx context.Context) error {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}
	startup, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	pool, err := db.Open(startup, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	store := db.Store{Pool: pool}
	objects, err := storage.New(cfg)
	if err != nil {
		return err
	}
	ready := func(ctx context.Context) error {
		if err := store.Check(ctx); err != nil {
			return err
		}
		return objects.Check(ctx)
	}
	handler := api.NewRouter(api.Dependencies{Store: store, Ready: ready, Logger: slog.Default(), Limiter: rate.NewLimiter(rate.Limit(cfg.RateLimit), cfg.RateBurst), Timeout: cfg.Timeout})
	listener, err := net.Listen("tcp", cfg.HTTPAddress)
	if err != nil {
		return errors.New("HTTP listener unavailable")
	}
	return server.Run(ctx, listener, handler, slog.Default())
}
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx); err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
}
