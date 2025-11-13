package usecases

import (
	"context"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type CreateTeamUseCase struct {
	storage   TeamStorage
	userStore UserStorage
}

func NewCreateTeamUseCase(teamStorage TeamStorage, userStorage UserStorage) *CreateTeamUseCase {
	return &CreateTeamUseCase{
		storage:   teamStorage,
		userStore: userStorage,
	}
}

func (uc *CreateTeamUseCase) Execute(ctx context.Context, team domain.Team) error {
	for i := range team.Users {
		user := team.Users[i]
		if !user.IsActive {
			user.IsActive = true
		}
		if err := uc.userStore.CreateUser(ctx, user); err != nil {
			return err
		}
		team.Users[i] = user
	}

	return uc.storage.CreateTeam(ctx, team)
}
