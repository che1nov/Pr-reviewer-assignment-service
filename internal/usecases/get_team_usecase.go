package usecases

import (
	"context"
	"log/slog"

	"github.com/che1nov/Pr-reviewer-assignment-service/internal/domain"
)

type GetTeamUseCase struct {
	teams TeamStorage
	log   *slog.Logger
}

func NewGetTeamUseCase(teamStorage TeamStorage, log *slog.Logger) *GetTeamUseCase {
	return &GetTeamUseCase{
		teams: teamStorage,
		log:   log,
	}
}

// Get находит команду.
func (uc *GetTeamUseCase) Get(ctx context.Context, name string) (domain.Team, error) {
	uc.log.InfoContext(ctx, "получаем команду", "team_name", name)

	team, err := uc.teams.GetTeam(ctx, name)
	if err != nil {
		uc.log.WarnContext(ctx, "команда не найдена", "team_name", name, "error", err)
		return domain.Team{}, err
	}

	return team, nil
}
