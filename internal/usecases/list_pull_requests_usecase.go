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

type GetReviewerPullRequestsUseCase struct {
	storage PullRequestStorage
}

func NewGetReviewerPullRequestsUseCase(storage PullRequestStorage) *GetReviewerPullRequestsUseCase {
	return &GetReviewerPullRequestsUseCase{storage: storage}
}

func (uc *GetReviewerPullRequestsUseCase) Execute(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	return uc.storage.ListPullRequestsByReviewer(ctx, reviewerID)
}
