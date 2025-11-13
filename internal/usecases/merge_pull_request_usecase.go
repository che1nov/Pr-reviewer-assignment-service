package usecases

import (
	"context"
	"log/slog"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

// MergePullRequestUseCase помечает pull request как MERGED.
type MergePullRequestUseCase struct {
	prs   PullRequestStorage
	clock ClockAdapter
	log   *slog.Logger
}

// NewMergePullRequestUseCase создаёт use case merge pull request.
func NewMergePullRequestUseCase(prStorage PullRequestStorage, clock ClockAdapter, log *slog.Logger) *MergePullRequestUseCase {
	return &MergePullRequestUseCase{
		prs:   prStorage,
		clock: clock,
		log:   log,
	}
}

// Execute выполняет merge, операция идемпотентна.
func (uc *MergePullRequestUseCase) Execute(ctx context.Context, id string) (domain.PullRequest, error) {
	uc.log.InfoContext(ctx, "merge pull request", "pr_id", id)

	pr, err := uc.prs.GetPullRequest(ctx, id)
	if err != nil {
		uc.log.WarnContext(ctx, "pull request не найден", "pr_id", id, "error", err)
		return domain.PullRequest{}, err
	}

	pr.MarkMerged(uc.clock.Now())

	if err := uc.prs.UpdatePullRequest(ctx, pr); err != nil {
		uc.log.ErrorContext(ctx, "не удалось обновить pull request", "pr_id", id, "error", err)
		return domain.PullRequest{}, err
	}

	uc.log.InfoContext(ctx, "pull request в статусе MERGED", "pr_id", id)
	return pr, nil
}
