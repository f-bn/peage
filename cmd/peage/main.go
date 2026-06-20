package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"peage/internal/config"
	"peage/internal/proxy"
)

// Version
var (
	version    = "dev"
	commitHash = "unknown"
	buildDate  = "unknown"
)

func main() {
	// Parse configuration
	cfg, err := config.Parse()
	if err != nil {
		slog.Error("Configuration failed", "error", err)
		os.Exit(1)
	}

	// Logging
	level := slog.LevelInfo
	if cfg.Verbose {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))

	// Create and start the listener server
	slog.Info("Starting Peage", "version", version, "commit", commitHash, "buildDate", buildDate)
	rp := proxy.New(cfg.SocketPath)
	handler := proxy.Handler(rp, cfg.Engine)

	server := &http.Server{
		Addr:    cfg.ListenAddress,
		Handler: handler,
	}

	go func() {
		slog.Info("Starting server", "address", cfg.ListenAddress, "socket", cfg.SocketPath, "engine", cfg.Engine)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Error during server startup", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("Shutting down server", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server failed to stopped gracefully", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped gracefully")
}
