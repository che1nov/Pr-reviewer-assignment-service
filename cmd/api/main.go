package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/config"
	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/app"
	"github.com/che1nov/backend-trainee-assignment-autumn-2025/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.New(logger.Config{Level: cfg.LogLevel})

	application := app.New(cfg, log)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*15))
		defer cancel()
		if err := application.Shutdown(shutdownCtx); err != nil {
			log.Error("shutdown failed", "error", err)
		}
	}()

	if err := application.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server error", "error", err)
	}
}
