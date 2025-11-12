package usecases

import (
	"context"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type UserStorage interface {
	CreateUser(ctx context.Context, id, name string) error
	ListUsers(ctx context.Context) ([]domain.User, error)
}
