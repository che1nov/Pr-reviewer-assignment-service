package usecases

import (
	"context"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type ReassignReviewerUseCase struct {
	prStorage   PullRequestStorage
	teamStorage TeamStorage
	userStorage UserStorage
}

func NewReassignReviewerUseCase(
	prStorage PullRequestStorage,
	teamStorage TeamStorage,
	userStorage UserStorage,
) *ReassignReviewerUseCase {
	return &ReassignReviewerUseCase{
		prStorage:   prStorage,
		teamStorage: teamStorage,
		userStorage: userStorage,
	}
}

func (uc *ReassignReviewerUseCase) Execute(ctx context.Context, prID, oldReviewerID, newReviewerID string) (domain.PullRequest, error) {
	pr, err := uc.prStorage.GetPullRequest(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, err
	}

	team, err := uc.teamStorage.GetTeam(ctx, pr.TeamName)
	if err != nil {
		return domain.PullRequest{}, err
	}

	if oldReviewerID == newReviewerID {
		return domain.PullRequest{}, domain.ErrReviewerAlreadyAdded
	}

	var newMember domain.User
	isNewMember := false
	for _, member := range team.Users {
		if member.ID == newReviewerID {
			newMember = member
			isNewMember = true
			break
		}
	}
	if !isNewMember {
		return domain.PullRequest{}, domain.ErrReviewerNotInTeam
	}
	if !newMember.IsActive {
		return domain.PullRequest{}, domain.ErrReviewerInactive
	}

	if err := pr.ReplaceReviewer(oldReviewerID, newReviewerID); err != nil {
		return domain.PullRequest{}, err
	}

	if err := uc.prStorage.UpdatePullRequest(ctx, pr); err != nil {
		return domain.PullRequest{}, err
	}

	return pr, nil
}
