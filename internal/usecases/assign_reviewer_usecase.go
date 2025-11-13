package usecases

import (
	"context"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type AssignReviewerUseCase struct {
	prStorage   PullRequestStorage
	teamStorage TeamStorage
}

func NewAssignReviewerUseCase(prStorage PullRequestStorage, teamStorage TeamStorage) *AssignReviewerUseCase {
	return &AssignReviewerUseCase{
		prStorage:   prStorage,
		teamStorage: teamStorage,
	}
}

func (uc *AssignReviewerUseCase) Execute(ctx context.Context, prID, userID string) (domain.PullRequest, error) {
	pr, err := uc.prStorage.GetPullRequest(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, err
	}

	team, err := uc.teamStorage.GetTeam(ctx, pr.TeamName)
	if err != nil {
		return domain.PullRequest{}, err
	}

	// проверяем, что пользователь есть в команде и не автор
	isTeamMember := false
	for _, member := range team.Users {
		if member.ID == userID {
			isTeamMember = true
			break
		}
	}
	if !isTeamMember {
		return domain.PullRequest{}, domain.ErrReviewerAlreadyAdded
	}

	if err := pr.AddReviewer(userID); err != nil {
		return domain.PullRequest{}, err
	}

	if err := uc.prStorage.UpdatePullRequest(ctx, pr); err != nil {
		return domain.PullRequest{}, err
	}

	return pr, nil
}
