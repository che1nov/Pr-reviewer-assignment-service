package usecases

import (
	"context"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type ListPullRequestsUseCase struct {
	storage PullRequestStorage
}

func NewListPullRequestsUseCase(storage PullRequestStorage) *ListPullRequestsUseCase {
	return &ListPullRequestsUseCase{storage: storage}
}

func (uc *ListPullRequestsUseCase) Execute(ctx context.Context) ([]domain.PullRequest, error) {
	return uc.storage.ListPullRequests(ctx)
}
