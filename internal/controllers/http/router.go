package httpcontroller

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/dto"
	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/usecases"
)

// RouterConfig содержит зависимости маршрутизатора
type RouterConfig struct {
	Message                  string
	CreateUserUseCase        *usecases.CreateUserUseCase
	ListUsersUseCase         *usecases.ListUsersUseCase
	SetUserActiveUseCase     *usecases.SetUserActiveUseCase
	CreateTeamUseCase        *usecases.CreateTeamUseCase
	ListTeamsUseCase         *usecases.ListTeamsUseCase
	CreatePullRequestUseCase *usecases.CreatePullRequestUseCase
	ListPullRequestsUseCase  *usecases.ListPullRequestsUseCase
	MergePullRequestUseCase  *usecases.MergePullRequestUseCase
	AssignReviewerUseCase    *usecases.AssignReviewerUseCase
	ReassignReviewerUseCase  *usecases.ReassignReviewerUseCase
	GetReviewerPRsUseCase    *usecases.GetReviewerPullRequestsUseCase
	Logger                   *slog.Logger
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cfg.Message))
	})

	if cfg.ListUsersUseCase != nil {
		r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
			users, err := cfg.ListUsersUseCase.Execute(r.Context())
			if err != nil {
				cfg.Logger.Error("failed to list users", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			response := dto.UsersOutput{
				Users: make([]dto.UserOutput, 0, len(users)),
			}
			for _, user := range users {
				response.Users = append(response.Users, dto.UserOutput{
					ID:       user.ID,
					Name:     user.Name,
					IsActive: user.IsActive,
				})
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				cfg.Logger.Error("failed to encode users", "error", err)
			}
		})
	}

	if cfg.CreateUserUseCase != nil {
		r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
			var input dto.UserInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				cfg.Logger.Error("failed to decode user", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if input.ID == "" || input.Name == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			isActive := true
			if input.IsActive != nil {
				isActive = *input.IsActive
			}

			user := domain.User{
				ID:       input.ID,
				Name:     input.Name,
				IsActive: isActive,
			}

			if err := cfg.CreateUserUseCase.Execute(r.Context(), user); err != nil {
				cfg.Logger.Error("failed to create user", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(dto.UserOutput{
				ID:       user.ID,
				Name:     user.Name,
				IsActive: user.IsActive,
			})
		})
	}

	if cfg.SetUserActiveUseCase != nil {
		r.Post("/users/{id}/status", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			if id == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var payload struct {
				IsActive bool `json:"is_active"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				cfg.Logger.Error("failed to decode user status payload", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			user, err := cfg.SetUserActiveUseCase.Execute(r.Context(), id, payload.IsActive)
			if err != nil {
				if errors.Is(err, domain.ErrUserNotFound) {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				cfg.Logger.Error("failed to update user status", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(dto.UserOutput{
				ID:       user.ID,
				Name:     user.Name,
				IsActive: user.IsActive,
			})
		})
	}

	if cfg.ListTeamsUseCase != nil {
		r.Get("/teams", func(w http.ResponseWriter, r *http.Request) {
			teams, err := cfg.ListTeamsUseCase.Execute(r.Context())
			if err != nil {
				cfg.Logger.Error("failed to list teams", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			response := make([]dto.TeamOutput, 0, len(teams))
			for _, team := range teams {
				users := make([]dto.UserOutput, 0, len(team.Users))
				for _, user := range team.Users {
					users = append(users, dto.UserOutput{
						ID:       user.ID,
						Name:     user.Name,
						IsActive: user.IsActive,
					})
				}
				response = append(response, dto.TeamOutput{
					Name:  team.Name,
					Users: users,
				})
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				cfg.Logger.Error("failed to encode teams", "error", err)
			}
		})
	}

	if cfg.CreateTeamUseCase != nil {
		r.Post("/teams", func(w http.ResponseWriter, r *http.Request) {
			var input dto.TeamInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				cfg.Logger.Error("failed to decode team", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			users := make([]domain.User, 0, len(input.Users))
			for _, u := range input.Users {
				if u.ID == "" || u.Name == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				isActive := true
				if u.IsActive != nil {
					isActive = *u.IsActive
				}
				users = append(users, domain.User{ID: u.ID, Name: u.Name, IsActive: isActive})
			}

			team := domain.NewTeam(input.Name, users)
			if err := cfg.CreateTeamUseCase.Execute(r.Context(), team); err != nil {
				cfg.Logger.Error("failed to create team", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
		})
	}

	if cfg.ListPullRequestsUseCase != nil {
		r.Get("/pull-requests", func(w http.ResponseWriter, r *http.Request) {
			prs, err := cfg.ListPullRequestsUseCase.Execute(r.Context())
			if err != nil {
				cfg.Logger.Error("failed to list pull requests", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			response := make([]dto.PullRequestOutput, 0, len(prs))
			for _, pr := range prs {
				response = append(response, dto.PullRequestOutput{
					ID:                pr.ID,
					Title:             pr.Title,
					AuthorID:          pr.AuthorID,
					TeamName:          pr.TeamName,
					Reviewers:         pr.Reviewers,
					Status:            pr.Status,
					NeedMoreReviewers: pr.NeedMoreReviewers,
				})
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				cfg.Logger.Error("failed to encode pull requests", "error", err)
			}
		})
	}

	if cfg.CreatePullRequestUseCase != nil {
		r.Post("/pull-requests", func(w http.ResponseWriter, r *http.Request) {
			var input dto.CreatePullRequestInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				cfg.Logger.Error("failed to decode pull request", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if input.ID == "" || input.Title == "" || input.AuthorID == "" || input.TeamName == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			pr := domain.NewPullRequest(input.ID, input.Title, input.AuthorID, input.TeamName)

			created, err := cfg.CreatePullRequestUseCase.Execute(r.Context(), pr)
			if err != nil {
				cfg.Logger.Error("failed to create pull request", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(dto.PullRequestOutput{
				ID:                created.ID,
				Title:             created.Title,
				AuthorID:          created.AuthorID,
				TeamName:          created.TeamName,
				Reviewers:         created.Reviewers,
				Status:            created.Status,
				NeedMoreReviewers: created.NeedMoreReviewers,
			})
		})
	}

	if cfg.MergePullRequestUseCase != nil {
		r.Post("/pull-requests/{id}/merge", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			if id == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			merged, err := cfg.MergePullRequestUseCase.Execute(r.Context(), id)
			if err != nil {
				if errors.Is(err, domain.ErrPullRequestNotFound) {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				cfg.Logger.Error("failed to merge pull request", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(dto.PullRequestOutput{
				ID:                merged.ID,
				Title:             merged.Title,
				AuthorID:          merged.AuthorID,
				TeamName:          merged.TeamName,
				Reviewers:         merged.Reviewers,
				Status:            merged.Status,
				NeedMoreReviewers: merged.NeedMoreReviewers,
			})
		})
	}

	if cfg.AssignReviewerUseCase != nil {
		r.Post("/pull-requests/{id}/reviewers", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			if id == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var payload struct {
				UserID string `json:"user_id"`
			}

			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				cfg.Logger.Error("failed to decode reviewer payload", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if payload.UserID == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			updated, err := cfg.AssignReviewerUseCase.Execute(r.Context(), id, payload.UserID)
			if err != nil {
				switch {
				case errors.Is(err, domain.ErrPullRequestNotFound):
					w.WriteHeader(http.StatusNotFound)
				case errors.Is(err, domain.ErrTeamNotFound):
					w.WriteHeader(http.StatusBadRequest)
				case errors.Is(err, domain.ErrReviewerAlreadyAdded):
					w.WriteHeader(http.StatusConflict)
				case errors.Is(err, domain.ErrPullRequestMerged):
					w.WriteHeader(http.StatusConflict)
				case errors.Is(err, domain.ErrReviewerInactive):
					w.WriteHeader(http.StatusConflict)
				case errors.Is(err, domain.ErrReviewerNotInTeam):
					w.WriteHeader(http.StatusBadRequest)
				case errors.Is(err, domain.ErrReviewerIsAuthor):
					w.WriteHeader(http.StatusConflict)
				case errors.Is(err, domain.ErrReviewerLimitReached):
					w.WriteHeader(http.StatusConflict)
				default:
					cfg.Logger.Error("failed to assign reviewer", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
				}
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(dto.PullRequestOutput{
				ID:                updated.ID,
				Title:             updated.Title,
				AuthorID:          updated.AuthorID,
				TeamName:          updated.TeamName,
				Reviewers:         updated.Reviewers,
				Status:            updated.Status,
				NeedMoreReviewers: updated.NeedMoreReviewers,
			})
		})
	}

	if cfg.ReassignReviewerUseCase != nil {
		r.Post("/pull-requests/{id}/reviewers/reassign", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			if id == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var payload struct {
				OldReviewerID string  `json:"old_reviewer_id"`
				NewReviewerID *string `json:"new_reviewer_id,omitempty"`
			}

			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				cfg.Logger.Error("failed to decode reassign payload", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if payload.OldReviewerID == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			newReviewerID := ""
			if payload.NewReviewerID != nil {
				newReviewerID = *payload.NewReviewerID
			}

			updated, err := cfg.ReassignReviewerUseCase.Execute(r.Context(), id, payload.OldReviewerID, newReviewerID)
			if err != nil {
				switch {
				case errors.Is(err, domain.ErrPullRequestNotFound):
					w.WriteHeader(http.StatusNotFound)
				case errors.Is(err, domain.ErrTeamNotFound):
					w.WriteHeader(http.StatusBadRequest)
				case errors.Is(err, domain.ErrReviewerNotAssigned):
					w.WriteHeader(http.StatusNotFound)
				case errors.Is(err, domain.ErrReviewerInactive):
					w.WriteHeader(http.StatusConflict)
				case errors.Is(err, domain.ErrReviewerNotInTeam):
					w.WriteHeader(http.StatusBadRequest)
				case errors.Is(err, domain.ErrReviewerIsAuthor):
					w.WriteHeader(http.StatusConflict)
				case errors.Is(err, domain.ErrReviewerAlreadyAdded):
					w.WriteHeader(http.StatusConflict)
				case errors.Is(err, domain.ErrPullRequestMerged):
					w.WriteHeader(http.StatusConflict)
				case errors.Is(err, domain.ErrReviewerLimitReached):
					w.WriteHeader(http.StatusConflict)
				default:
					cfg.Logger.Error("failed to reassign reviewer", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
				}
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(dto.PullRequestOutput{
				ID:                updated.ID,
				Title:             updated.Title,
				AuthorID:          updated.AuthorID,
				TeamName:          updated.TeamName,
				Reviewers:         updated.Reviewers,
				Status:            updated.Status,
				NeedMoreReviewers: updated.NeedMoreReviewers,
			})
		})
	}

	if cfg.GetReviewerPRsUseCase != nil {
		r.Get("/users/{id}/reviews", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			if id == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			prs, err := cfg.GetReviewerPRsUseCase.Execute(r.Context(), id)
			if err != nil {
				cfg.Logger.Error("failed to list reviewer pull requests", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			response := dto.ReviewerPullRequestsOutput{
				ReviewerID:   id,
				PullRequests: make([]dto.PullRequestOutput, 0, len(prs)),
			}
			for _, pr := range prs {
				response.PullRequests = append(response.PullRequests, dto.PullRequestOutput{
					ID:                pr.ID,
					Title:             pr.Title,
					AuthorID:          pr.AuthorID,
					TeamName:          pr.TeamName,
					Reviewers:         pr.Reviewers,
					Status:            pr.Status,
					NeedMoreReviewers: pr.NeedMoreReviewers,
				})
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				cfg.Logger.Error("failed to encode reviewer pull requests", "error", err)
			}
		})
	}

	return r
}
