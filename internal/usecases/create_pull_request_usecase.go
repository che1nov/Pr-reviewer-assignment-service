package usecases

import (
	"context"
	"errors"
	"log/slog"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type CreatePullRequestUseCase struct {
	prs   PullRequestStorage
	teams TeamStorage
	users UserStorage
	clock ClockAdapter
	rand  RandomAdapter
	log   *slog.Logger
}

func NewCreatePullRequestUseCase(
	prStorage PullRequestStorage,
	teamStorage TeamStorage,
	userStorage UserStorage,
	clock ClockAdapter,
	random RandomAdapter,
	log *slog.Logger,
) *CreatePullRequestUseCase {
	return &CreatePullRequestUseCase{
		prs:   prStorage,
		teams: teamStorage,
		users: userStorage,
		clock: clock,
		rand:  random,
		log:   log,
	}
}

// Create создаёт pull request.
func (uc *CreatePullRequestUseCase) Create(ctx context.Context, id, title, authorID string) (domain.PullRequest, error) {
	uc.log.InfoContext(ctx, "создаём pull request", "pr_id", id, "author_id", authorID)

	if _, err := uc.prs.GetPullRequest(ctx, id); err == nil {
		uc.log.WarnContext(ctx, "pull request уже существует", "pr_id", id)
		return domain.PullRequest{}, domain.ErrPullRequestExists
	} else if !errors.Is(err, domain.ErrPullRequestNotFound) {
		uc.log.ErrorContext(ctx, "ошибка проверки существующего pull request", "error", err, "pr_id", id)
		return domain.PullRequest{}, err
	}

	author, err := uc.users.GetUser(ctx, authorID)
	if err != nil {
		uc.log.WarnContext(ctx, "автор не найден", "author_id", authorID, "error", err)
		return domain.PullRequest{}, err
	}
	if author.TeamName == "" {
		uc.log.WarnContext(ctx, "у автора нет команды", "author_id", authorID)
		return domain.PullRequest{}, domain.ErrTeamNotFound
	}

	team, err := uc.teams.GetTeam(ctx, author.TeamName)
	if err != nil {
		uc.log.WarnContext(ctx, "команда автора не найдена", "team_name", author.TeamName, "error", err)
		return domain.PullRequest{}, err
	}

	candidates := team.ActiveReviewersExcluding(authorID)
	if len(candidates) == 0 {
		uc.log.WarnContext(ctx, "нет активных кандидатов в ревьюеры", "pr_id", id, "team_name", team.Name)
	}

	selected := candidates
	if len(selected) > 0 {
		if uc.rand != nil && len(selected) > 1 {
			uc.rand.Shuffle(len(selected), func(i, j int) {
				selected[i], selected[j] = selected[j], selected[i]
			})
		}

		if len(selected) > 2 {
			selected = selected[:2]
		}
	}

	reviewerIDs := make([]string, 0, len(selected))
	for _, reviewer := range selected {
		reviewerIDs = append(reviewerIDs, reviewer.ID)
	}

	pr := domain.NewPullRequest(id, title, authorID, team.Name, uc.clock.Now())
	pr.AssignReviewers(reviewerIDs)

	if err := uc.prs.CreatePullRequest(ctx, pr); err != nil {
		uc.log.ErrorContext(ctx, "ошибка сохранения pull request", "error", err, "pr_id", id)
		return domain.PullRequest{}, err
	}

	uc.log.InfoContext(ctx, "pull request создан", "pr_id", id, "reviewers", reviewerIDs)
	return pr, nil
}
