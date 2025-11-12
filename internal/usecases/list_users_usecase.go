package usecases

import (
	"context"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

// ListUsersUseCase возвращает всех пользователей из хранилища
type ListUsersUseCase struct {
	storage UserStorage
}

func NewListUsersUseCase(storage UserStorage) *ListUsersUseCase {
	return &ListUsersUseCase{storage: storage}
}

func (uc *ListUsersUseCase) Execute(ctx context.Context) ([]domain.User, error) {
	return uc.storage.ListUsers(ctx)
}
