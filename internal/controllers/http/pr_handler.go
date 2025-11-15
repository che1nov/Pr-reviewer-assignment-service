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

type PullRequestHandler struct {
	logger            *slog.Logger
	createPRUseCase   *usecases.CreatePullRequestUseCase
	mergePRUseCase    *usecases.MergePullRequestUseCase
	reassignPRUseCase *usecases.ReassignReviewerUseCase
}

func NewPullRequestHandler(
	logger *slog.Logger,
	createPRUseCase *usecases.CreatePullRequestUseCase,
	mergePRUseCase *usecases.MergePullRequestUseCase,
	reassignPRUseCase *usecases.ReassignReviewerUseCase,
) *PullRequestHandler {
	return &PullRequestHandler{
		logger:            logger,
		createPRUseCase:   createPRUseCase,
		mergePRUseCase:    mergePRUseCase,
		reassignPRUseCase: reassignPRUseCase,
	}
}

// Create создаёт новый pull request.
func (h *PullRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body dto.CreatePullRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondBadRequest(h.logger, r, w, "BAD_REQUEST", "некорректный формат запроса", err)
		return
	}
	if body.PullRequestID == "" || body.PullRequestName == "" || body.AuthorID == "" {
		respondBadRequest(h.logger, r, w, "BAD_REQUEST", "pull_request_id, pull_request_name и author_id обязательны", nil)
		return
	}

	pr, err := h.createPRUseCase.Create(r.Context(), body.PullRequestID, body.PullRequestName, body.AuthorID)
	if err != nil {
		status, code, message := mapCreatePRError(err)
		h.logger.ErrorContext(r.Context(), "ошибка создания pull request", "error", err, "pr_id", body.PullRequestID)
		respondError(h.logger, w, status, code, message)
		return
	}

	respondJSON(h.logger, w, http.StatusCreated, map[string]dto.PullRequest{"pr": toPullRequest(pr)})
}

// Merge помечает pull request как MERGED.
func (h *PullRequestHandler) Merge(w http.ResponseWriter, r *http.Request) {
	var body dto.MergePullRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondBadRequest(h.logger, r, w, "BAD_REQUEST", "некорректный формат запроса", err)
		return
	}
	if body.PullRequestID == "" {
		respondBadRequest(h.logger, r, w, "BAD_REQUEST", "pull_request_id обязателен", nil)
		return
	}

	pr, err := h.mergePRUseCase.Merge(r.Context(), body.PullRequestID)
	if err != nil {
		status, code, message := mapMergePRError(err)
		h.logger.ErrorContext(r.Context(), "ошибка merge pull request", "error", err, "pr_id", body.PullRequestID)
		respondError(h.logger, w, status, code, message)
		return
	}

	respondJSON(h.logger, w, http.StatusOK, map[string]dto.PullRequest{"pr": toPullRequest(pr)})
}

// Reassign заменяет ревьюера в pull request.
func (h *PullRequestHandler) Reassign(w http.ResponseWriter, r *http.Request) {
	var body dto.ReassignReviewerRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondBadRequest(h.logger, r, w, "BAD_REQUEST", "некорректный формат запроса", err)
		return
	}
	if body.PullRequestID == "" || body.OldReviewerID == "" {
		respondBadRequest(h.logger, r, w, "BAD_REQUEST", "pull_request_id и old_user_id обязательны", nil)
		return
	}

	pr, replacedBy, err := h.reassignPRUseCase.Reassign(r.Context(), body.PullRequestID, body.OldReviewerID, body.NewReviewerID)
	if err != nil {
		status, code, message := mapReassignPRError(err)
		h.logger.ErrorContext(r.Context(), "ошибка переназначения ревьюера", "error", err, "pr_id", body.PullRequestID, "old_reviewer", body.OldReviewerID)
		respondError(h.logger, w, status, code, message)
		return
	}

	respondJSON(h.logger, w, http.StatusOK, dto.ReassignReviewerResponse{
		PR:         toPullRequest(pr),
		ReplacedBy: replacedBy,
	})
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

func mapCreatePRError(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrPullRequestExists):
		return http.StatusConflict, ErrCodePRExists, "pull request already exists"
	case errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound, ErrCodeNotFound, "author not found"
	case errors.Is(err, domain.ErrTeamNotFound):
		return http.StatusNotFound, ErrCodeNotFound, "team not found"
	case errors.Is(err, domain.ErrNoReviewerCandidates):
		return http.StatusConflict, ErrCodeNoCandidate, "no active reviewer candidates in team"
	default:
		return http.StatusInternalServerError, ErrCodeInternal, "internal error"
	}
}

func mapMergePRError(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrPullRequestNotFound):
		return http.StatusNotFound, ErrCodeNotFound, "pull request not found"
	default:
		return http.StatusInternalServerError, ErrCodeInternal, "internal error"
	}
}

func mapReassignPRError(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrPullRequestNotFound):
		return http.StatusNotFound, ErrCodeNotFound, "pull request not found"
	case errors.Is(err, domain.ErrReviewerNotAssigned):
		return http.StatusConflict, ErrCodeNotAssigned, "reviewer is not assigned to this PR"
	case errors.Is(err, domain.ErrPullRequestMerged):
		return http.StatusConflict, ErrCodePRMerged, "cannot reassign reviewer on merged PR"
	case errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound, ErrCodeNotFound, "user not found"
	case errors.Is(err, domain.ErrNoReviewerCandidates):
		return http.StatusConflict, ErrCodeNoCandidate, "no active replacement candidate in team"
	case errors.Is(err, domain.ErrReviewerInactive):
		return http.StatusConflict, ErrCodeNoCandidate, "reviewer inactive"
	default:
		return http.StatusInternalServerError, ErrCodeInternal, "internal error"
	}
}
