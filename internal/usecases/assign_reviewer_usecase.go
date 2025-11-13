package usecases

import (
	"context"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type AssignReviewerUseCase struct {
	prStorage   PullRequestStorage
	teamStorage TeamStorage
	userStorage UserStorage
}

func NewAssignReviewerUseCase(prStorage PullRequestStorage, teamStorage TeamStorage, userStorage UserStorage) *AssignReviewerUseCase {
	return &AssignReviewerUseCase{
		prStorage:   prStorage,
		teamStorage: teamStorage,
		userStorage: userStorage,
	}
}

func (uc *AssignReviewerUseCase) Execute(ctx context.Context, prID, userID string) (domain.PullRequest, error) {
	pr, err := uc.prStorage.GetPullRequest(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, err
	}

	member, err := uc.userStorage.GetUser(ctx, userID)
	if err != nil {
		return domain.PullRequest{}, err
	}
	if !member.IsActive {
		return domain.PullRequest{}, domain.ErrReviewerInactive
	}

	team, err := uc.teamStorage.GetTeam(ctx, pr.TeamName)
	if err != nil {
		return domain.PullRequest{}, err
	}

	isTeamMember := false
	for _, teamMember := range team.Users {
		if teamMember.ID == userID {
			isTeamMember = true
			break
		}
	}
	if !isTeamMember {
		return domain.PullRequest{}, domain.ErrReviewerNotInTeam
	}

	if err := pr.AddReviewer(userID); err != nil {
		return domain.PullRequest{}, err
	}

	if err := uc.prStorage.UpdatePullRequest(ctx, pr); err != nil {
		return domain.PullRequest{}, err
	}

	return pr, nil
}
