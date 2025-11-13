package httpcontroller

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/dto"
	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/usecases"
)

type TeamHandler struct {
	logger    *slog.Logger
	addTeamUC *usecases.CreateTeamUseCase
	getTeamUC *usecases.GetTeamUseCase
}

func NewTeamHandler(logger *slog.Logger, addTeamUC *usecases.CreateTeamUseCase, getTeamUC *usecases.GetTeamUseCase) *TeamHandler {
	return &TeamHandler{
		logger:    logger,
		addTeamUC: addTeamUC,
		getTeamUC: getTeamUC,
	}
}

// AddTeam обрабатывает создание команды.
func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var body dto.Team
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondBadRequest(h.logger, r, w, "BAD_REQUEST", "некорректный формат запроса", err)
		return
	}

	if body.TeamName == "" {
		respondBadRequest(h.logger, r, w, "BAD_REQUEST", "team_name обязателен", nil)
		return
	}

	members := make([]domain.User, 0, len(body.Members))
	for _, member := range body.Members {
		if member.UserID == "" || member.Username == "" {
			respondBadRequest(h.logger, r, w, "BAD_REQUEST", "user_id и username обязательны", nil)
			return
		}
		members = append(members, domain.NewUser(member.UserID, member.Username, body.TeamName, member.IsActive))
	}

	team, err := h.addTeamUC.Create(r.Context(), domain.NewTeam(body.TeamName, members))
	if err != nil {
		status, code, message := mapTeamError(err)
		h.logger.ErrorContext(r.Context(), "ошибка создания команды", "error", err, "team_name", body.TeamName)
		respondError(h.logger, w, status, code, message)
		return
	}

	respondJSON(h.logger, w, http.StatusCreated, map[string]dto.Team{"team": toTeam(team)})
}

// GetTeam возвращает данные команды.
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		respondBadRequest(h.logger, r, w, "BAD_REQUEST", "team_name обязателен", nil)
		return
	}

	team, err := h.getTeamUC.Get(r.Context(), teamName)
	if err != nil {
		h.logger.WarnContext(r.Context(), "команда не найдена", "team_name", teamName, "error", err)
		respondError(h.logger, w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	respondJSON(h.logger, w, http.StatusOK, toTeam(team))
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

func mapTeamError(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrTeamExists):
		return http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists"
	default:
		return http.StatusInternalServerError, "INTERNAL", "internal error"
	}
}
