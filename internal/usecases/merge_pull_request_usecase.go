package usecases

import (
	"context"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type MergePullRequestUseCase struct {
	prStorage PullRequestStorage
}

func NewMergePullRequestUseCase(prStorage PullRequestStorage) *MergePullRequestUseCase {
	return &MergePullRequestUseCase{prStorage: prStorage}
}

func (uc *MergePullRequestUseCase) Execute(ctx context.Context, id string) (domain.PullRequest, error) {
	pr, err := uc.prStorage.GetPullRequest(ctx, id)
	if err != nil {
		return domain.PullRequest{}, err
	}

	pr.MarkMerged()

	if err := uc.prStorage.UpdatePullRequest(ctx, pr); err != nil {
		return domain.PullRequest{}, err
	}

	return pr, nil
}
