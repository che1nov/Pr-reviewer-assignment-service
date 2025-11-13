package usecases

import (
	"context"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type SetUserActiveUseCase struct {
	storage UserStorage
}

func NewSetUserActiveUseCase(storage UserStorage) *SetUserActiveUseCase {
	return &SetUserActiveUseCase{storage: storage}
}

func (uc *SetUserActiveUseCase) Execute(ctx context.Context, id string, isActive bool) (domain.User, error) {
	user, err := uc.storage.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	user.IsActive = isActive
	if err := uc.storage.UpdateUser(ctx, user); err != nil {
		return domain.User{}, err
	}

	return user, nil
}
