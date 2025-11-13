package usecases

import (
	"context"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type CreatePullRequestUseCase struct {
	storage PullRequestStorage
}

func NewCreatePullRequestUseCase(storage PullRequestStorage) *CreatePullRequestUseCase {
	return &CreatePullRequestUseCase{storage: storage}
}

func (uc *CreatePullRequestUseCase) Execute(ctx context.Context, pr domain.PullRequest) error {
	return uc.storage.CreatePullRequest(ctx, pr)
}
