package usecases

import (
	"context"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type ListTeamsUseCase struct {
	storage TeamStorage
}

func NewListTeamsUseCase(storage TeamStorage) *ListTeamsUseCase {
	return &ListTeamsUseCase{storage: storage}
}

func (uc *ListTeamsUseCase) Execute(ctx context.Context) ([]domain.Team, error) {
	return uc.storage.ListTeams(ctx)
}
