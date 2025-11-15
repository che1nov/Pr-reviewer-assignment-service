package usecases

import (
	"context"
	"log/slog"

	"github.com/che1nov/Pr-reviewer-assignment-service/internal/domain"
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

	team, err := uc.teams.GetTeam(ctx, teamName)
	if err != nil {
		uc.log.WarnContext(ctx, "команда не найдена", "team_name", teamName, "error", err)
		return DeactivateResult{}, err
	}

	if len(team.Users) == 0 {
		uc.log.InfoContext(ctx, "команда пустая, ничего не делаем", "team_name", teamName)
		return DeactivateResult{DeactivatedCount: 0, ReassignedPRCount: 0}, nil
	}

	userIDs := uc.getActiveUserIDs(team)
	if len(userIDs) == 0 {
		uc.log.InfoContext(ctx, "все пользователи уже неактивны", "team_name", teamName)
		return DeactivateResult{DeactivatedCount: 0, ReassignedPRCount: 0}, nil
	}

	affectedPRs, userIDMap, err := uc.filterAffectedPRs(ctx, userIDs)
	if err != nil {
		return DeactivateResult{}, err
	}

	reassignedCount, err := uc.reassignAffectedPRs(ctx, affectedPRs, userIDMap)
	if err != nil {
		return DeactivateResult{}, err
	}

	deactivatedCount, err := uc.deactivateUsers(ctx, userIDs)
	if err != nil {
		return DeactivateResult{}, err
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

// getActiveUserIDs собирает ID активных пользователей команды
func (uc *DeactivateTeamUsersUseCase) getActiveUserIDs(team domain.Team) []string {
	userIDs := make([]string, 0, len(team.Users))
	for _, user := range team.Users {
		if user.IsActive {
			userIDs = append(userIDs, user.ID)
		}
	}
	return userIDs
}

// filterAffectedPRs фильтрует открытые PR с ревьюверами из команды
func (uc *DeactivateTeamUsersUseCase) filterAffectedPRs(ctx context.Context, userIDs []string) ([]domain.PullRequest, map[string]bool, error) {
	allPRs, err := uc.prs.ListPullRequests(ctx)
	if err != nil {
		uc.log.ErrorContext(ctx, "ошибка получения PR", "error", err)
		return nil, nil, err
	}

	userIDMap := make(map[string]bool, len(userIDs))
	for _, id := range userIDs {
		userIDMap[id] = true
	}

	affectedPRs := make([]domain.PullRequest, 0)
	for _, pr := range allPRs {
		if pr.Status != domain.PRStatusOpen {
			continue
		}
		if uc.hasAffectedReviewer(pr, userIDMap) {
			affectedPRs = append(affectedPRs, pr)
		}
	}

	uc.log.InfoContext(ctx, "найдены затронутые PR", "count", len(affectedPRs))
	return affectedPRs, userIDMap, nil
}

// hasAffectedReviewer проверяет есть ли среди ревьюверов PR деактивируемые пользователи
func (uc *DeactivateTeamUsersUseCase) hasAffectedReviewer(pr domain.PullRequest, userIDMap map[string]bool) bool {
	for _, reviewerID := range pr.Reviewers {
		if userIDMap[reviewerID] {
			return true
		}
	}
	return false
}

// reassignAffectedPRs переназначает ревьюверов во всех затронутых PR
func (uc *DeactivateTeamUsersUseCase) reassignAffectedPRs(ctx context.Context, affectedPRs []domain.PullRequest, userIDMap map[string]bool) (int, error) {
	reassignedCount := 0
	for _, pr := range affectedPRs {
		updated, err := uc.reassignReviewersForPR(ctx, pr, userIDMap)
		if err != nil {
			return 0, err
		}
		if updated {
			reassignedCount++
		}
	}
	return reassignedCount, nil
}

// reassignReviewersForPR переназначает ревьюверов для одного PR
func (uc *DeactivateTeamUsersUseCase) reassignReviewersForPR(ctx context.Context, pr domain.PullRequest, userIDMap map[string]bool) (bool, error) {
	authorTeam, err := uc.teams.GetTeam(ctx, pr.TeamName)
	if err != nil {
		uc.log.WarnContext(ctx, "не найдена команда автора PR", "pr_id", pr.ID, "team", pr.TeamName)
		return false, nil
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

	if !updated {
		return false, nil
	}

	pr.Reviewers = newReviewers
	if err := uc.prs.UpdatePullRequest(ctx, pr); err != nil {
		uc.log.ErrorContext(ctx, "ошибка обновления PR", "pr_id", pr.ID, "error", err)
		return false, err
	}

	return true, nil
}

// deactivateUsers деактивирует всех указанных пользователей
func (uc *DeactivateTeamUsersUseCase) deactivateUsers(ctx context.Context, userIDs []string) (int, error) {
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
			return 0, err
		}
		deactivatedCount++
	}
	return deactivatedCount, nil
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
