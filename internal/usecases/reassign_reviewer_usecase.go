package usecases

import (
	"context"
	"errors"
	"log/slog"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

// ReassignReviewerUseCase переназначает ревьюера на pull request.
type ReassignReviewerUseCase struct {
	prs   PullRequestStorage
	teams TeamStorage
	users UserStorage
	rand  RandomAdapter
	log   *slog.Logger
}

// NewReassignReviewerUseCase создаёт use case переназначения ревьюера.
func NewReassignReviewerUseCase(
	prStorage PullRequestStorage,
	teamStorage TeamStorage,
	userStorage UserStorage,
	random RandomAdapter,
	log *slog.Logger,
) *ReassignReviewerUseCase {
	return &ReassignReviewerUseCase{
		prs:   prStorage,
		teams: teamStorage,
		users: userStorage,
		rand:  random,
		log:   log,
	}
}

// Execute переназначает ревьюера. Возвращает новый состав и идентификатор заменяющего.
func (uc *ReassignReviewerUseCase) Execute(ctx context.Context, prID, oldReviewerID string, desiredNew *string) (domain.PullRequest, string, error) {
	uc.log.InfoContext(ctx, "переназначаем ревьюера", "pr_id", prID, "old_reviewer", oldReviewerID)

	pr, err := uc.prs.GetPullRequest(ctx, prID)
	if err != nil {
		uc.log.WarnContext(ctx, "pull request не найден", "pr_id", prID, "error", err)
		return domain.PullRequest{}, "", err
	}

	team, err := uc.teams.GetTeam(ctx, pr.TeamName)
	if err != nil {
		uc.log.WarnContext(ctx, "команда не найдена при переназначении", "team_name", pr.TeamName, "error", err)
		return domain.PullRequest{}, "", err
	}

	candidates := make([]domain.User, 0, len(team.Users))
	for _, member := range team.Users {
		if member.ID == pr.AuthorID || !member.IsActive {
			continue
		}
		if member.ID == oldReviewerID {
			continue
		}
		if contains(pr.Reviewers, member.ID) {
			continue
		}
		candidates = append(candidates, member)
	}

	var newReviewerID string
	if desiredNew != nil && *desiredNew != "" {
		uc.log.InfoContext(ctx, "используем указанного нового ревьюера", "candidate_id", *desiredNew)
		if !containsUser(candidates, *desiredNew) {
			uc.log.WarnContext(ctx, "указанный кандидат недоступен", "candidate_id", *desiredNew)
			return domain.PullRequest{}, "", domain.ErrNoReviewerCandidates
		}
		newReviewerID = *desiredNew
	} else {
		if len(candidates) == 0 {
			uc.log.WarnContext(ctx, "нет доступных кандидатов для переназначения", "pr_id", prID)
			return domain.PullRequest{}, "", domain.ErrNoReviewerCandidates
		}
		if uc.rand != nil && len(candidates) > 1 {
			uc.rand.Shuffle(len(candidates), func(i, j int) {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			})
		}
		newReviewerID = candidates[0].ID
	}

	if _, err := uc.users.GetUser(ctx, newReviewerID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			uc.log.WarnContext(ctx, "новый ревьюер не найден", "user_id", newReviewerID)
			return domain.PullRequest{}, "", err
		}
		uc.log.ErrorContext(ctx, "ошибка получения нового ревьюера", "user_id", newReviewerID, "error", err)
		return domain.PullRequest{}, "", err
	}

	if err := pr.ReplaceReviewer(oldReviewerID, newReviewerID); err != nil {
		uc.log.WarnContext(ctx, "ошибка ReplaceReviewer", "error", err, "pr_id", prID)
		return domain.PullRequest{}, "", err
	}

	if err := uc.prs.UpdatePullRequest(ctx, pr); err != nil {
		uc.log.ErrorContext(ctx, "не удалось сохранить pull request", "error", err, "pr_id", prID)
		return domain.PullRequest{}, "", err
	}

	uc.log.InfoContext(ctx, "переназначение выполнено", "pr_id", prID, "new_reviewer", newReviewerID)
	return pr, newReviewerID, nil
}

func contains(list []string, target string) bool {
	for _, value := range list {
		if value == target {
			return true
		}
	}
	return false
}

func containsUser(list []domain.User, target string) bool {
	for _, user := range list {
		if user.ID == target {
			return true
		}
	}
	return false
}
