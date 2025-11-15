package usecases

import (
	"context"
	"log/slog"

	"github.com/che1nov/Pr-reviewer-assignment-service/internal/domain"
)

type ListPullRequestsUseCase struct {
	prs PullRequestStorage
	log *slog.Logger
}

func NewListPullRequestsUseCase(storage PullRequestStorage, log *slog.Logger) *ListPullRequestsUseCase {
	return &ListPullRequestsUseCase{
		prs: storage,
		log: log,
	}
}

// List возвращает все pull request.
func (uc *ListPullRequestsUseCase) List(ctx context.Context) ([]domain.PullRequest, error) {
	uc.log.InfoContext(ctx, "получаем список pull request")

	prs, err := uc.prs.ListPullRequests(ctx)
	if err != nil {
		uc.log.ErrorContext(ctx, "ошибка выборки pull request", "error", err)
		return nil, err
	}

	return prs, nil
}
