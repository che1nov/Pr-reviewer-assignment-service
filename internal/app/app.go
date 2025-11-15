package app

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/che1nov/Pr-reviewer-assignment-service/config"
	"github.com/che1nov/Pr-reviewer-assignment-service/internal/adapters/postgresql"
	httpcontroller "github.com/che1nov/Pr-reviewer-assignment-service/internal/controllers/http"
	"github.com/che1nov/Pr-reviewer-assignment-service/internal/usecases"
	"github.com/che1nov/Pr-reviewer-assignment-service/pkg/clock"
	"github.com/che1nov/Pr-reviewer-assignment-service/pkg/random"
)

// App запущенное приложение
type App struct {
	server *http.Server
	logger *slog.Logger
	cfg    config.Config
	db     *sqlx.DB
}

// New настройка приложения.
func New(cfg config.Config, logger *slog.Logger) (*App, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL обязателен для работы сервиса")
	}

	connection, err := postgresql.NewConnection(cfg.DatabaseURL, logger)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	if err := postgresql.RunMigrations(connection.DB, logger); err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("ошибка применения миграций: %w", err)
	}

	db := connection
	userStorage := postgresql.NewUserAdapter(connection, logger)
	teamStorage := postgresql.NewTeamAdapter(connection, logger)
	prStorage := postgresql.NewPullRequestAdapter(connection, logger)

	clockAdapter := clock.NewSystem()
	randomAdapter := random.New(rand.New(rand.NewSource(time.Now().UnixNano())))

	createTeamUC := usecases.NewCreateTeamUseCase(teamStorage, userStorage, logger)
	getTeamUC := usecases.NewGetTeamUseCase(teamStorage, logger)
	setUserActiveUC := usecases.NewSetUserActiveUseCase(userStorage, logger)
	createPullRequestUC := usecases.NewCreatePullRequestUseCase(prStorage, teamStorage, userStorage, clockAdapter, randomAdapter, logger)
	mergePullRequestUC := usecases.NewMergePullRequestUseCase(prStorage, clockAdapter, logger)
	reassignReviewerUC := usecases.NewReassignReviewerUseCase(prStorage, teamStorage, userStorage, randomAdapter, logger)
	getReviewerPRsUC := usecases.NewGetReviewerPullRequestsUseCase(prStorage, logger)
	getStatsUC := usecases.NewGetStatsUseCase(prStorage, userStorage, logger)
	deactivateTeamUsersUC := usecases.NewDeactivateTeamUsersUseCase(userStorage, teamStorage, prStorage, randomAdapter, logger)

	router := httpcontroller.NewRouter(httpcontroller.RouterConfig{
		Logger:                     logger,
		AdminToken:                 cfg.AdminToken,
		UserToken:                  cfg.UserToken,
		AddTeamUseCase:             createTeamUC,
		GetTeamUseCase:             getTeamUC,
		SetUserActiveUseCase:       setUserActiveUC,
		CreatePullRequestUseCase:   createPullRequestUC,
		MergePullRequestUseCase:    mergePullRequestUC,
		ReassignReviewerUseCase:    reassignReviewerUC,
		GetReviewerPRsUseCase:      getReviewerPRsUC,
		GetStatsUseCase:            getStatsUC,
		DeactivateTeamUsersUseCase: deactivateTeamUsersUC,
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
		db:     db,
	}, nil
}

// Handler для тестирования.
func (a *App) Handler() http.Handler {
	return a.server.Handler
}

// Start запускает HTTP сервер.
func (a *App) Start() error {
	a.logger.Info("запускаем HTTP сервер", "addr", a.server.Addr)
	return a.server.ListenAndServe()
}

// Shutdown останавливает сервер.
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("останавливаем HTTP сервер")
	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			a.logger.Warn("ошибка закрытия соединения с PostgreSQL", "error", err)
		}
	}

	return nil
}
