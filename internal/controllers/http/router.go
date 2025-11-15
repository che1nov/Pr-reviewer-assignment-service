package httpcontroller

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/usecases"
)

type RouterConfig struct {
	Logger *slog.Logger

	AdminToken string
	UserToken  string

	AddTeamUseCase             *usecases.CreateTeamUseCase
	GetTeamUseCase             *usecases.GetTeamUseCase
	SetUserActiveUseCase       *usecases.SetUserActiveUseCase
	CreatePullRequestUseCase   *usecases.CreatePullRequestUseCase
	MergePullRequestUseCase    *usecases.MergePullRequestUseCase
	ReassignReviewerUseCase    *usecases.ReassignReviewerUseCase
	GetReviewerPRsUseCase      *usecases.GetReviewerPullRequestsUseCase
	GetStatsUseCase            *usecases.GetStatsUseCase
	DeactivateTeamUsersUseCase *usecases.DeactivateTeamUsersUseCase
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(cfg.Logger, w, http.StatusOK, map[string]string{"status": "ok"})
	})

	teamHandler := NewTeamHandler(cfg.Logger, cfg.AddTeamUseCase, cfg.GetTeamUseCase)
	prHandler := NewPullRequestHandler(cfg.Logger, cfg.CreatePullRequestUseCase, cfg.MergePullRequestUseCase, cfg.ReassignReviewerUseCase)
	userHandler := NewUserHandler(cfg.Logger, cfg.SetUserActiveUseCase, cfg.GetReviewerPRsUseCase)
	statsHandler := NewStatsHandler(cfg.Logger, cfg.GetStatsUseCase)
	deactivateHandler := NewDeactivateHandler(cfg.Logger, cfg.DeactivateTeamUsersUseCase)

	r.Group(func(admin chi.Router) {
		admin.Use(adminAuth(cfg.Logger, cfg.AdminToken))

		admin.Post("/team/add", teamHandler.AddTeam)
		admin.Post("/team/deactivateUsers", deactivateHandler.DeactivateTeamUsers)
		admin.Post("/pullRequest/create", prHandler.Create)
		admin.Post("/pullRequest/merge", prHandler.Merge)
		admin.Post("/pullRequest/reassign", prHandler.Reassign)
		admin.Post("/users/setIsActive", userHandler.SetActive)
	})

	r.Group(func(user chi.Router) {
		user.Use(userAuth(cfg.Logger, cfg.AdminToken, cfg.UserToken))

		user.Get("/team/get", teamHandler.GetTeam)
		user.Get("/users/getReview", userHandler.GetReviews)
		user.Get("/stats", statsHandler.GetStats)
	})

	return r
}
