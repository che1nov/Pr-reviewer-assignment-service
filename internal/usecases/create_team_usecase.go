package usecases

import (
	"context"
	"errors"
	"log/slog"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type CreateTeamUseCase struct {
	teams TeamStorage
	users UserStorage
	log   *slog.Logger
}

func NewCreateTeamUseCase(teamStorage TeamStorage, userStorage UserStorage, log *slog.Logger) *CreateTeamUseCase {
	return &CreateTeamUseCase{
		teams: teamStorage,
		users: userStorage,
		log:   log,
	}
}

// Create создаёт новую команду и обновляет участников.
func (uc *CreateTeamUseCase) Create(ctx context.Context, team domain.Team) (domain.Team, error) {
	uc.log.InfoContext(ctx, "создаём команду", "team_name", team.Name)

	if _, err := uc.teams.GetTeam(ctx, team.Name); err == nil {
		uc.log.WarnContext(ctx, "команда уже существует", "team_name", team.Name)
		return domain.Team{}, domain.ErrTeamExists
	} else if !errors.Is(err, domain.ErrTeamNotFound) {
		uc.log.ErrorContext(ctx, "ошибка чтения команды", "error", err, "team_name", team.Name)
		return domain.Team{}, err
	}

	for i := range team.Users {
		member := team.Users[i]
		member.TeamName = team.Name

		existing, err := uc.users.GetUser(ctx, member.ID)
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			if err := uc.users.CreateUser(ctx, member); err != nil {
				uc.log.ErrorContext(ctx, "не удалось создать пользователя", "error", err, "user_id", member.ID)
				return domain.Team{}, err
			}
			uc.log.InfoContext(ctx, "создали нового пользователя", "user_id", member.ID, "team_name", team.Name)
		case err != nil:
			uc.log.ErrorContext(ctx, "не удалось получить пользователя", "error", err, "user_id", member.ID)
			return domain.Team{}, err
		default:
			existing.Name = member.Name
			existing.TeamName = team.Name
			existing.IsActive = member.IsActive
			if err := uc.users.UpdateUser(ctx, existing); err != nil {
				uc.log.ErrorContext(ctx, "не удалось обновить пользователя", "error", err, "user_id", member.ID)
				return domain.Team{}, err
			}
			member = existing
			uc.log.InfoContext(ctx, "обновили пользователя", "user_id", member.ID, "team_name", team.Name)
		}

		team.Users[i] = member
	}

	if err := uc.teams.CreateTeam(ctx, team); err != nil {
		uc.log.ErrorContext(ctx, "не удалось сохранить команду", "error", err, "team_name", team.Name)
		return domain.Team{}, err
	}

	uc.log.InfoContext(ctx, "команда успешно создана", "team_name", team.Name)
	return team, nil
}
