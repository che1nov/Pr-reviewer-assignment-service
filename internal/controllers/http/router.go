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

// RouterConfig зависимости HTTP-роутера.
type RouterConfig struct {
	Logger *slog.Logger

	AdminToken string
	UserToken  string

	AddTeamUseCase           *usecases.CreateTeamUseCase
	GetTeamUseCase           *usecases.GetTeamUseCase
	SetUserActiveUseCase     *usecases.SetUserActiveUseCase
	CreatePullRequestUseCase *usecases.CreatePullRequestUseCase
	MergePullRequestUseCase  *usecases.MergePullRequestUseCase
	ReassignReviewerUseCase  *usecases.ReassignReviewerUseCase
	GetReviewerPRsUseCase    *usecases.GetReviewerPullRequestsUseCase
}

// NewRouter строит HTTP маршрутизатор в соответствии с OpenAPI.
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(cfg.Logger, w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Group(func(r chi.Router) {
		r.Use(adminAuth(cfg.Logger, cfg.AdminToken))

		r.Post("/team/add", func(w http.ResponseWriter, r *http.Request) {
			var body dto.Team
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				respondBadRequest(cfg.Logger, r, w, "BAD_REQUEST", "некорректный формат запроса", err)
				return
			}

			if body.TeamName == "" {
				respondBadRequest(cfg.Logger, r, w, "BAD_REQUEST", "team_name обязателен", nil)
				return
			}

			members := make([]domain.User, 0, len(body.Members))
			for _, member := range body.Members {
				if member.UserID == "" || member.Username == "" {
					respondBadRequest(cfg.Logger, r, w, "BAD_REQUEST", "user_id и username обязательны", nil)
					return
				}
				members = append(members, domain.NewUser(member.UserID, member.Username, body.TeamName, member.IsActive))
			}

			team, err := cfg.AddTeamUseCase.Execute(r.Context(), domain.NewTeam(body.TeamName, members))
			if err != nil {
				status, code, message := mapTeamError(err)
				cfg.Logger.ErrorContext(r.Context(), "ошибка создания команды", "error", err, "team_name", body.TeamName)
				respondError(cfg.Logger, w, status, code, message)
				return
			}

			respondJSON(cfg.Logger, w, http.StatusCreated, map[string]dto.Team{"team": toTeam(team)})
		})

		r.Post("/pullRequest/create", func(w http.ResponseWriter, r *http.Request) {
			var body dto.CreatePullRequestRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				respondBadRequest(cfg.Logger, r, w, "BAD_REQUEST", "некорректный формат запроса", err)
				return
			}
			if body.PullRequestID == "" || body.PullRequestName == "" || body.AuthorID == "" {
				respondBadRequest(cfg.Logger, r, w, "BAD_REQUEST", "pull_request_id, pull_request_name и author_id обязательны", nil)
				return
			}

			pr, err := cfg.CreatePullRequestUseCase.Execute(r.Context(), body.PullRequestID, body.PullRequestName, body.AuthorID)
			if err != nil {
				status, code, message := mapPullRequestCreateError(err)
				cfg.Logger.ErrorContext(r.Context(), "ошибка создания pull request", "error", err, "pr_id", body.PullRequestID)
				respondError(cfg.Logger, w, status, code, message)
				return
			}

			respondJSON(cfg.Logger, w, http.StatusCreated, map[string]dto.PullRequest{"pr": toPullRequest(pr)})
		})

		r.Post("/pullRequest/merge", func(w http.ResponseWriter, r *http.Request) {
			var body dto.MergePullRequestRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				respondBadRequest(cfg.Logger, r, w, "BAD_REQUEST", "некорректный формат запроса", err)
				return
			}
			if body.PullRequestID == "" {
				respondBadRequest(cfg.Logger, r, w, "BAD_REQUEST", "pull_request_id обязателен", nil)
				return
			}

			pr, err := cfg.MergePullRequestUseCase.Execute(r.Context(), body.PullRequestID)
			if err != nil {
				status, code, message := mapMergeError(err)
				cfg.Logger.ErrorContext(r.Context(), "ошибка при merge pull request", "error", err, "pr_id", body.PullRequestID)
				respondError(cfg.Logger, w, status, code, message)
				return
			}

			respondJSON(cfg.Logger, w, http.StatusOK, map[string]dto.PullRequest{"pr": toPullRequest(pr)})
		})

		r.Post("/pullRequest/reassign", func(w http.ResponseWriter, r *http.Request) {
			var body dto.ReassignReviewerRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				respondBadRequest(cfg.Logger, r, w, "BAD_REQUEST", "некорректный формат запроса", err)
				return
			}
			if body.PullRequestID == "" || body.OldReviewerID == "" {
				respondBadRequest(cfg.Logger, r, w, "BAD_REQUEST", "pull_request_id и old_user_id обязательны", nil)
				return
			}

			pr, replacedBy, err := cfg.ReassignReviewerUseCase.Execute(r.Context(), body.PullRequestID, body.OldReviewerID, body.NewReviewerID)
			if err != nil {
				status, code, message := mapReassignError(err)
				cfg.Logger.ErrorContext(r.Context(), "ошибка переназначения ревьюера", "error", err, "pr_id", body.PullRequestID, "old_reviewer", body.OldReviewerID)
				respondError(cfg.Logger, w, status, code, message)
				return
			}

			respondJSON(cfg.Logger, w, http.StatusOK, dto.ReassignReviewerResponse{
				PR:         toPullRequest(pr),
				ReplacedBy: replacedBy,
			})
		})

		r.Post("/users/setIsActive", func(w http.ResponseWriter, r *http.Request) {
			var body dto.SetUserActiveRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				respondBadRequest(cfg.Logger, r, w, "BAD_REQUEST", "некорректный формат запроса", err)
				return
			}
			if body.UserID == "" {
				respondBadRequest(cfg.Logger, r, w, "BAD_REQUEST", "user_id обязателен", nil)
				return
			}

			user, err := cfg.SetUserActiveUseCase.Execute(r.Context(), body.UserID, body.IsActive)
			if err != nil {
				status, code, message := mapUserError(err)
				cfg.Logger.ErrorContext(r.Context(), "ошибка обновления статуса пользователя", "error", err, "user_id", body.UserID)
				respondError(cfg.Logger, w, status, code, message)
				return
			}

			respondJSON(cfg.Logger, w, http.StatusOK, map[string]dto.User{"user": toUser(user)})
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(userAuth(cfg.Logger, cfg.AdminToken, cfg.UserToken))

		r.Get("/team/get", func(w http.ResponseWriter, r *http.Request) {
			teamName := r.URL.Query().Get("team_name")
			if teamName == "" {
				respondBadRequest(cfg.Logger, r, w, "BAD_REQUEST", "team_name обязателен", nil)
				return
			}

			team, err := cfg.GetTeamUseCase.Execute(r.Context(), teamName)
			if err != nil {
				cfg.Logger.WarnContext(r.Context(), "команда не найдена", "team_name", teamName, "error", err)
				respondError(cfg.Logger, w, http.StatusNotFound, "NOT_FOUND", err.Error())
				return
			}

			respondJSON(cfg.Logger, w, http.StatusOK, toTeam(team))
		})

		r.Get("/users/getReview", func(w http.ResponseWriter, r *http.Request) {
			userID := r.URL.Query().Get("user_id")
			if userID == "" {
				respondBadRequest(cfg.Logger, r, w, "BAD_REQUEST", "user_id обязателен", nil)
				return
			}

			prs, err := cfg.GetReviewerPRsUseCase.Execute(r.Context(), userID)
			if err != nil {
				cfg.Logger.ErrorContext(r.Context(), "ошибка получения pull request пользователя", "error", err, "user_id", userID)
				respondError(cfg.Logger, w, http.StatusInternalServerError, "INTERNAL", err.Error())
				return
			}

			response := dto.ReviewerPullRequestsResponse{
				UserID:       userID,
				PullRequests: make([]dto.PullRequestShort, 0, len(prs)),
			}
			for _, pr := range prs {
				response.PullRequests = append(response.PullRequests, dto.PullRequestShort{
					PullRequestID:   pr.ID,
					PullRequestName: pr.Title,
					AuthorID:        pr.AuthorID,
					Status:          pr.Status,
				})
			}

			respondJSON(cfg.Logger, w, http.StatusOK, response)
		})
	})

	return r
}

func adminAuth(logger *slog.Logger, token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !checkBearer(r, token) {
				logger.WarnContext(r.Context(), "неверный админ токен")
				respondError(logger, w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func userAuth(logger *slog.Logger, adminToken, userToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if checkBearer(r, adminToken) || checkBearer(r, userToken) {
				next.ServeHTTP(w, r)
				return
			}
			logger.WarnContext(r.Context(), "неверный пользовательский токен")
			respondError(logger, w, http.StatusUnauthorized, "UNAUTHORIZED", "token required")
		})
	}
}

func checkBearer(r *http.Request, token string) bool {
	if token == "" {
		return false
	}
	value := r.Header.Get("Authorization")
	const prefix = "Bearer "
	return len(value) == len(prefix)+len(token) && value[:len(prefix)] == prefix && value[len(prefix):] == token
}

func respondJSON(logger *slog.Logger, w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error("не удалось закодировать ответ", "error", err)
	}
}

func respondError(logger *slog.Logger, w http.ResponseWriter, status int, code, message string) {
	respondJSON(logger, w, status, dto.ErrorResponse{
		Error: dto.ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func respondBadRequest(logger *slog.Logger, r *http.Request, w http.ResponseWriter, code, message string, err error) {
	if err != nil {
		logger.WarnContext(r.Context(), "ошибка декодирования запроса", "error", err)
	} else {
		logger.WarnContext(r.Context(), "ошибка валидации запроса", "message", message)
	}
	respondError(logger, w, http.StatusBadRequest, code, message)
}

func toTeam(team domain.Team) dto.Team {
	result := dto.Team{
		TeamName: team.Name,
		Members:  make([]dto.TeamMember, 0, len(team.Users)),
	}
	for _, user := range team.Users {
		result.Members = append(result.Members, dto.TeamMember{
			UserID:   user.ID,
			Username: user.Name,
			IsActive: user.IsActive,
		})
	}
	return result
}

func toUser(user domain.User) dto.User {
	return dto.User{
		UserID:   user.ID,
		Username: user.Name,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}

func toPullRequest(pr domain.PullRequest) dto.PullRequest {
	return dto.PullRequest{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Title,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status,
		AssignedReviewers: append([]string(nil), pr.Reviewers...),
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func mapTeamError(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrTeamExists):
		return http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists"
	default:
		return http.StatusInternalServerError, "INTERNAL", "internal error"
	}
}

func mapUserError(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound, "NOT_FOUND", "user not found"
	default:
		return http.StatusInternalServerError, "INTERNAL", "internal error"
	}
}

func mapPullRequestCreateError(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrPullRequestExists):
		return http.StatusConflict, "PR_EXISTS", "pull request already exists"
	case errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound, "NOT_FOUND", "author not found"
	case errors.Is(err, domain.ErrTeamNotFound):
		return http.StatusNotFound, "NOT_FOUND", "team not found"
	case errors.Is(err, domain.ErrNoReviewerCandidates):
		return http.StatusConflict, "NO_CANDIDATE", "no active reviewer candidates in team"
	default:
		return http.StatusInternalServerError, "INTERNAL", "internal error"
	}
}

func mapMergeError(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrPullRequestNotFound):
		return http.StatusNotFound, "NOT_FOUND", "pull request not found"
	default:
		return http.StatusInternalServerError, "INTERNAL", "internal error"
	}
}

func mapReassignError(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrPullRequestNotFound):
		return http.StatusNotFound, "NOT_FOUND", "pull request not found"
	case errors.Is(err, domain.ErrReviewerNotAssigned):
		return http.StatusConflict, "NOT_ASSIGNED", "reviewer is not assigned to this PR"
	case errors.Is(err, domain.ErrPullRequestMerged):
		return http.StatusConflict, "PR_MERGED", "cannot reassign reviewer on merged PR"
	case errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound, "NOT_FOUND", "user not found"
	case errors.Is(err, domain.ErrNoReviewerCandidates):
		return http.StatusConflict, "NO_CANDIDATE", "no active replacement candidate in team"
	case errors.Is(err, domain.ErrReviewerInactive):
		return http.StatusConflict, "NO_CANDIDATE", "reviewer inactive"
	default:
		return http.StatusInternalServerError, "INTERNAL", "internal error"
	}
}
