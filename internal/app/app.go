package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/config"
	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/adapters/memory"
	httpcontroller "github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/controllers/http"
	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/usecases"
)

type App struct {
	server *http.Server
	logger *slog.Logger
	cfg    config.Config
}

func New(cfg config.Config, logger *slog.Logger) *App {
	userStorage := memory.NewUserStorage()
	createUserUC := usecases.NewCreateUserUseCase(userStorage)
	listUsersUC := usecases.NewListUsersUseCase(userStorage)
	createTeamUC := usecases.NewCreateTeamUseCase(userStorage, userStorage)
	listTeamsUC := usecases.NewListTeamsUseCase(userStorage)
	createPRUC := usecases.NewCreatePullRequestUseCase(userStorage, userStorage)
	listPRsUC := usecases.NewListPullRequestsUseCase(userStorage)
	mergePRUC := usecases.NewMergePullRequestUseCase(userStorage)
	assignReviewerUC := usecases.NewAssignReviewerUseCase(userStorage, userStorage)

	router := httpcontroller.NewRouter(httpcontroller.RouterConfig{
		Message:                  cfg.Message,
		CreateUserUseCase:        createUserUC,
		ListUsersUseCase:         listUsersUC,
		CreateTeamUseCase:        createTeamUC,
		ListTeamsUseCase:         listTeamsUC,
		CreatePullRequestUseCase: createPRUC,
		ListPullRequestsUseCase:  listPRsUC,
		MergePullRequestUseCase:  mergePRUC,
		AssignReviewerUseCase:    assignReviewerUC,
		Logger:                   logger,
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

func (a *App) Start() error {
	a.logger.Info("запускаем HTTP сервер", "addr", a.server.Addr)
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("останавливаем HTTP сервер")
	return a.server.Shutdown(ctx)
}

func (a *App) ShutdownTimeout() time.Duration {
	return a.cfg.ShutdownTimeout
}
