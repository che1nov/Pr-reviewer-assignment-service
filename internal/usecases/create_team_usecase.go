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
	if err := uc.storage.CreateTeam(ctx, team); err != nil {
		return err
	}

	for _, user := range team.Users {
		if err := uc.userStore.CreateUser(ctx, user.ID, user.Name); err != nil {
			return err
		}
	}

	return nil
}
