package usecases

import (
	"context"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type CreateUserUseCase struct {
	storage UserStorage
}

func NewCreateUserUseCase(storage UserStorage) *CreateUserUseCase {
	return &CreateUserUseCase{storage: storage}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, user domain.User) error {
	if !user.IsActive {
		user.IsActive = true
	}
	return uc.storage.CreateUser(ctx, user)
}
