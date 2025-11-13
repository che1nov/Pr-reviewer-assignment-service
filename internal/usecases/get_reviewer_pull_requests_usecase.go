package usecases

import (
	"context"
	"log/slog"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type GetReviewerPullRequestsUseCase struct {
	prs PullRequestStorage
	log *slog.Logger
}

func NewGetReviewerPullRequestsUseCase(storage PullRequestStorage, log *slog.Logger) *GetReviewerPullRequestsUseCase {
	return &GetReviewerPullRequestsUseCase{
		prs: storage,
		log: log,
	}
}

// ListByReviewer отдаёт pull request пользователя.
func (uc *GetReviewerPullRequestsUseCase) ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	uc.log.InfoContext(ctx, "получаем pull request для ревьюера", "reviewer_id", reviewerID)

	prs, err := uc.prs.ListPullRequestsByReviewer(ctx, reviewerID)
	if err != nil {
		uc.log.ErrorContext(ctx, "ошибка выборки pull request", "error", err, "reviewer_id", reviewerID)
		return nil, err
	}

	return prs, nil
}

