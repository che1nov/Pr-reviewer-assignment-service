package httpcontroller

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/che1nov/Pr-reviewer-assignment-service/internal/domain"
	"github.com/che1nov/Pr-reviewer-assignment-service/internal/dto"
	"github.com/che1nov/Pr-reviewer-assignment-service/internal/usecases"
)

type UserHandler struct {
	logger            *slog.Logger
	setActiveUseCase  *usecases.SetUserActiveUseCase
	getReviewsUseCase *usecases.GetReviewerPullRequestsUseCase
}

func NewUserHandler(
	logger *slog.Logger,
	setActiveUseCase *usecases.SetUserActiveUseCase,
	getReviewsUseCase *usecases.GetReviewerPullRequestsUseCase,
) *UserHandler {
	return &UserHandler{
		logger:            logger,
		setActiveUseCase:  setActiveUseCase,
		getReviewsUseCase: getReviewsUseCase,
	}
}

// SetActive обновляет флаг активности пользователя.
func (h *UserHandler) SetActive(w http.ResponseWriter, r *http.Request) {
	var body dto.SetUserActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondBadRequest(h.logger, r, w, "BAD_REQUEST", "некорректный формат запроса", err)
		return
	}
	if body.UserID == "" {
		respondBadRequest(h.logger, r, w, "BAD_REQUEST", "user_id обязателен", nil)
		return
	}

	user, err := h.setActiveUseCase.SetActive(r.Context(), body.UserID, body.IsActive)
	if err != nil {
		status, code, message := mapUserError(err)
		h.logger.ErrorContext(r.Context(), "ошибка обновления статуса пользователя", "error", err, "user_id", body.UserID)
		respondError(h.logger, w, status, code, message)
		return
	}

	respondJSON(h.logger, w, http.StatusOK, map[string]dto.User{"user": toUser(user)})
}

// GetReviews возвращает PR, где пользователь ревьювер.
func (h *UserHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		respondBadRequest(h.logger, r, w, "BAD_REQUEST", "user_id обязателен", nil)
		return
	}

	prs, err := h.getReviewsUseCase.ListByReviewer(r.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "ошибка получения pull request пользователя", "error", err, "user_id", userID)
		respondError(h.logger, w, http.StatusInternalServerError, "INTERNAL", err.Error())
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

	respondJSON(h.logger, w, http.StatusOK, response)
}

func toUser(user domain.User) dto.User {
	return dto.User{
		UserID:   user.ID,
		Username: user.Name,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
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
