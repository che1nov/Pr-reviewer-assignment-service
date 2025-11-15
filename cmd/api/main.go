package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/che1nov/Pr-reviewer-assignment-service/config"
	"github.com/che1nov/Pr-reviewer-assignment-service/internal/app"
	"github.com/che1nov/Pr-reviewer-assignment-service/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.New(logger.Config{Level: cfg.LogLevel})

	application, err := app.New(cfg, log)
	if err != nil {
		log.Error("failed to init application", "error", err)
		return
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := application.Shutdown(shutdownCtx); err != nil {
			log.Error("shutdown failed", "error", err)
		}
	}()

	if err := application.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server error", "error", err)
	}
}
