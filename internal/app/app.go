package app

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/config"
	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/adapters/memory"
	httpcontroller "github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/controllers/http"
	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/usecases"
	"github.com/che1nov/backend-trainee-assignment-autumn-2025/pkg/clock"
	"github.com/che1nov/backend-trainee-assignment-autumn-2025/pkg/random"
)

// App запущенное приложение
type App struct {
	server *http.Server
	logger *slog.Logger
	cfg    config.Config
}

// New настройка приложения
func New(cfg config.Config, logger *slog.Logger) *App {
	storage := memory.NewUserStorage(logger)
	clockAdapter := clock.NewSystem()
	randomAdapter := random.New(rand.New(rand.NewSource(time.Now().UnixNano())))

	createTeamUC := usecases.NewCreateTeamUseCase(storage, storage, logger)
	getTeamUC := usecases.NewGetTeamUseCase(storage, logger)
	setUserActiveUC := usecases.NewSetUserActiveUseCase(storage, logger)
	createPullRequestUC := usecases.NewCreatePullRequestUseCase(storage, storage, storage, clockAdapter, randomAdapter, logger)
	mergePullRequestUC := usecases.NewMergePullRequestUseCase(storage, clockAdapter, logger)
	reassignReviewerUC := usecases.NewReassignReviewerUseCase(storage, storage, storage, randomAdapter, logger)
	getReviewerPRsUC := usecases.NewGetReviewerPullRequestsUseCase(storage, logger)

	router := httpcontroller.NewRouter(httpcontroller.RouterConfig{
		Logger:                   logger,
		AdminToken:               cfg.AdminToken,
		UserToken:                cfg.UserToken,
		AddTeamUseCase:           createTeamUC,
		GetTeamUseCase:           getTeamUC,
		SetUserActiveUseCase:     setUserActiveUC,
		CreatePullRequestUseCase: createPullRequestUC,
		MergePullRequestUseCase:  mergePullRequestUC,
		ReassignReviewerUseCase:  reassignReviewerUC,
		GetReviewerPRsUseCase:    getReviewerPRsUC,
	})

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		server: server,
		logger: logger,
		cfg:    cfg,
	}
}

// Start запускает HTTP сервер.
func (a *App) Start() error {
	a.logger.Info("запускаем HTTP сервер", "addr", a.server.Addr)
	return a.server.ListenAndServe()
}

// Shutdown останавливает сервер.
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("останавливаем HTTP сервер")
	return a.server.Shutdown(ctx)
}
