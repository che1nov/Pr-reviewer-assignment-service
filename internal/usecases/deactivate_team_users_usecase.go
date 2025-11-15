package usecases

import (
	"context"
	"log/slog"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type DeactivateResult struct {
	DeactivatedCount  int `json:"deactivated_count"`
	ReassignedPRCount int `json:"reassigned_pr_count"`
}

type DeactivateTeamUsersUseCase struct {
	users  UserStorage
	teams  TeamStorage
	prs    PullRequestStorage
	random RandomAdapter
	log    *slog.Logger
}

func NewDeactivateTeamUsersUseCase(
	userStorage UserStorage,
	teamStorage TeamStorage,
	prStorage PullRequestStorage,
	random RandomAdapter,
	log *slog.Logger,
) *DeactivateTeamUsersUseCase {
	return &DeactivateTeamUsersUseCase{
		users:  userStorage,
		teams:  teamStorage,
		prs:    prStorage,
		random: random,
		log:    log,
	}
}

// DeactivateTeamUsers деактивирует всех пользователей команды и переназначает их открытые PR
func (uc *DeactivateTeamUsersUseCase) DeactivateTeamUsers(ctx context.Context, teamName string) (DeactivateResult, error) {
	uc.log.InfoContext(ctx, "массовая деактивация пользователей команды", "team_name", teamName)

	// Получаем команду
	team, err := uc.teams.GetTeam(ctx, teamName)
	if err != nil {
		uc.log.WarnContext(ctx, "команда не найдена", "team_name", teamName, "error", err)
		return DeactivateResult{}, err
	}

	if len(team.Users) == 0 {
		uc.log.InfoContext(ctx, "команда пустая, ничего не делаем", "team_name", teamName)
		return DeactivateResult{DeactivatedCount: 0, ReassignedPRCount: 0}, nil
	}

	// Собираем ID пользователей команды
	userIDs := make([]string, 0, len(team.Users))
	for _, user := range team.Users {
		if user.IsActive {
			userIDs = append(userIDs, user.ID)
		}
	}

	if len(userIDs) == 0 {
		uc.log.InfoContext(ctx, "все пользователи уже неактивны", "team_name", teamName)
		return DeactivateResult{DeactivatedCount: 0, ReassignedPRCount: 0}, nil
	}

	// Получаем все открытые PR где эти пользователи - ревьюверы
	allPRs, err := uc.prs.ListPullRequests(ctx)
	if err != nil {
		uc.log.ErrorContext(ctx, "ошибка получения PR", "error", err)
		return DeactivateResult{}, err
	}

	// Фильтруем открытые PR с ревьюверами из команды
	affectedPRs := make([]domain.PullRequest, 0)
	userIDMap := make(map[string]bool)
	for _, id := range userIDs {
		userIDMap[id] = true
	}

	for _, pr := range allPRs {
		if pr.Status != "OPEN" {
			continue
		}
		hasAffectedReviewer := false
		for _, reviewerID := range pr.Reviewers {
			if userIDMap[reviewerID] {
				hasAffectedReviewer = true
				break
			}
		}
		if hasAffectedReviewer {
			affectedPRs = append(affectedPRs, pr)
		}
	}

	uc.log.InfoContext(ctx, "найдены затронутые PR", "count", len(affectedPRs))

	reassignedCount := 0
	for _, pr := range affectedPRs {
		authorTeam, err := uc.teams.GetTeam(ctx, pr.TeamName)
		if err != nil {
			uc.log.WarnContext(ctx, "не найдена команда автора PR", "pr_id", pr.ID, "team", pr.TeamName)
			continue
		}

		newReviewers := make([]string, 0, len(pr.Reviewers))
		updated := false

		for _, reviewerID := range pr.Reviewers {
			if !userIDMap[reviewerID] {
				newReviewers = append(newReviewers, reviewerID)
				continue
			}

			replacement, found := uc.findReplacement(pr, reviewerID, authorTeam.Users, newReviewers)
			if !found {
				uc.log.WarnContext(ctx, "не найдена замена для ревьювера", "pr_id", pr.ID, "reviewer_id", reviewerID)
				updated = true
				continue
			}

			newReviewers = append(newReviewers, replacement)
			updated = true
			uc.log.InfoContext(ctx, "ревьювер заменен", "pr_id", pr.ID, "old", reviewerID, "new", replacement)
		}

		if updated {
			pr.Reviewers = newReviewers
			if err := uc.prs.UpdatePullRequest(ctx, pr); err != nil {
				uc.log.ErrorContext(ctx, "ошибка обновления PR", "pr_id", pr.ID, "error", err)
				return DeactivateResult{}, err
			}
			reassignedCount++
		}
	}

	deactivatedCount := 0
	for _, userID := range userIDs {
		user, err := uc.users.GetUser(ctx, userID)
		if err != nil {
			uc.log.WarnContext(ctx, "не найден пользователь", "user_id", userID, "error", err)
			continue
		}

		if !user.IsActive {
			continue
		}

		user.IsActive = false
		if err := uc.users.UpdateUser(ctx, user); err != nil {
			uc.log.ErrorContext(ctx, "ошибка деактивации пользователя", "user_id", userID, "error", err)
			return DeactivateResult{}, err
		}
		deactivatedCount++
	}

	uc.log.InfoContext(ctx, "массовая деактивация завершена",
		"team", teamName,
		"deactivated", deactivatedCount,
		"reassigned_prs", reassignedCount,
	)

	return DeactivateResult{
		DeactivatedCount:  deactivatedCount,
		ReassignedPRCount: reassignedCount,
	}, nil
}

// findReplacement ищет подходящую замену для ревьювера
func (uc *DeactivateTeamUsersUseCase) findReplacement(pr domain.PullRequest, oldReviewerID string, teamMembers []domain.User, currentReviewers []string) (string, bool) {
	candidates := make([]string, 0)

	for _, member := range teamMembers {
		if member.ID == pr.AuthorID {
			continue
		}
		if !member.IsActive {
			continue
		}
		if member.ID == oldReviewerID {
			continue
		}
		alreadyAssigned := false
		for _, existingReviewer := range currentReviewers {
			if existingReviewer == member.ID {
				alreadyAssigned = true
				break
			}
		}
		if alreadyAssigned {
			continue
		}

		candidates = append(candidates, member.ID)
	}

	if len(candidates) == 0 {
		return "", false
	}

	if uc.random != nil && len(candidates) > 1 {
		uc.random.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})
	}

	return candidates[0], true
}
