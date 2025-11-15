package usecases

import (
	"context"
	"log/slog"

	"github.com/che1nov/Pr-reviewer-assignment-service/internal/domain"
	"github.com/che1nov/Pr-reviewer-assignment-service/internal/dto"
)

type GetStatsUseCase struct {
	prs   PullRequestStorage
	users UserStorage
	log   *slog.Logger
}

func NewGetStatsUseCase(
	prStorage PullRequestStorage,
	userStorage UserStorage,
	log *slog.Logger,
) *GetStatsUseCase {
	return &GetStatsUseCase{
		prs:   prStorage,
		users: userStorage,
		log:   log,
	}
}

// GetStats возвращает статистику по назначениям ревьюверов
func (uc *GetStatsUseCase) GetStats(ctx context.Context) (dto.StatsResponse, error) {
	uc.log.InfoContext(ctx, "получаем статистику")

	allPRs, err := uc.prs.ListPullRequests(ctx)
	if err != nil {
		uc.log.ErrorContext(ctx, "ошибка получения списка PR", "error", err)
		return dto.StatsResponse{}, err
	}

	allUsers, err := uc.users.ListUsers(ctx)
	if err != nil {
		uc.log.ErrorContext(ctx, "ошибка получения списка пользователей", "error", err)
		return dto.StatsResponse{}, err
	}

	prStats := calculatePRStats(allPRs)

	userStats := calculateUserStats(allUsers, allPRs)

	uc.log.InfoContext(ctx, "статистика получена", "total_prs", prStats.TotalPRs, "users", len(userStats))

	return dto.StatsResponse{
		PRStats:   prStats,
		UserStats: userStats,
	}, nil
}

func calculatePRStats(prs []domain.PullRequest) dto.PRStats {
	stats := dto.PRStats{
		TotalPRs: len(prs),
	}

	reviewersMap := make(map[string]bool)

	for _, pr := range prs {
		if pr.Status == "OPEN" {
			stats.OpenPRs++
		} else if pr.Status == "MERGED" {
			stats.MergedPRs++
		}

		for _, reviewerID := range pr.Reviewers {
			reviewersMap[reviewerID] = true
		}
	}

	stats.ReviewersCount = len(reviewersMap)

	return stats
}

func calculateUserStats(users []domain.User, prs []domain.PullRequest) []dto.UserStats {
	statsMap := make(map[string]*dto.UserStats)

	for _, user := range users {
		statsMap[user.ID] = &dto.UserStats{
			UserID:      user.ID,
			Username:    user.Name,
			TeamName:    user.TeamName,
			AssignedPRs: 0,
			OpenPRs:     0,
			MergedPRs:   0,
		}
	}

	for _, pr := range prs {
		for _, reviewerID := range pr.Reviewers {
			if stats, exists := statsMap[reviewerID]; exists {
				stats.AssignedPRs++
				if pr.Status == "OPEN" {
					stats.OpenPRs++
				} else if pr.Status == "MERGED" {
					stats.MergedPRs++
				}
			}
		}
	}

	result := make([]dto.UserStats, 0)
	for _, stats := range statsMap {
		if stats.AssignedPRs > 0 {
			result = append(result, *stats)
		}
	}

	return result
}
