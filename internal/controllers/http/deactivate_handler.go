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

type DeactivateHandler struct {
	logger                     *slog.Logger
	deactivateTeamUsersUseCase *usecases.DeactivateTeamUsersUseCase
}

func NewDeactivateHandler(
	logger *slog.Logger,
	deactivateTeamUsersUseCase *usecases.DeactivateTeamUsersUseCase,
) *DeactivateHandler {
	return &DeactivateHandler{
		logger:                     logger,
		deactivateTeamUsersUseCase: deactivateTeamUsersUseCase,
	}
}

// DeactivateTeamUsers массово деактивирует пользователей команды
func (h *DeactivateHandler) DeactivateTeamUsers(w http.ResponseWriter, r *http.Request) {
	var input dto.DeactivateTeamUsersInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.WarnContext(r.Context(), "ошибка декодирования запроса", "error", err)
		respondError(h.logger, w, http.StatusBadRequest, "INVALID_INPUT", "неверный формат запроса")
		return
	}

	if input.TeamName == "" {
		respondError(h.logger, w, http.StatusBadRequest, "INVALID_INPUT", "team_name обязателен")
		return
	}

	result, err := h.deactivateTeamUsersUseCase.DeactivateTeamUsers(r.Context(), input.TeamName)
	if err != nil {
		status, code, message := mapDeactivateError(err)
		h.logger.ErrorContext(r.Context(), "ошибка массовой деактивации", "error", err, "team", input.TeamName)
		respondError(h.logger, w, status, code, message)
		return
	}

	output := dto.DeactivateTeamUsersOutput{
		DeactivatedCount:  result.DeactivatedCount,
		ReassignedPRCount: result.ReassignedPRCount,
	}

	respondJSON(h.logger, w, http.StatusOK, output)
}

func mapDeactivateError(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrTeamNotFound):
		return http.StatusNotFound, "NOT_FOUND", "team not found"
	default:
		return http.StatusInternalServerError, "INTERNAL", "internal error"
	}
}
