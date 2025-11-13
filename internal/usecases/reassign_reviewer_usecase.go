package usecases

import (
	"context"
	"math/rand"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type ReassignReviewerUseCase struct {
	prStorage   PullRequestStorage
	teamStorage TeamStorage
	userStorage UserStorage
	rand        *rand.Rand
}

func NewReassignReviewerUseCase(
	prStorage PullRequestStorage,
	teamStorage TeamStorage,
	userStorage UserStorage,
	rng *rand.Rand,
) *ReassignReviewerUseCase {
	return &ReassignReviewerUseCase{
		prStorage:   prStorage,
		teamStorage: teamStorage,
		userStorage: userStorage,
		rand:        rng,
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

	candidates := make([]domain.User, 0, len(team.Users))
	for _, member := range team.Users {
		if member.ID == pr.AuthorID || !member.IsActive {
			continue
		}
		if member.ID == oldReviewerID {
			continue
		}
		alreadyAssigned := false
		for _, assigned := range pr.Reviewers {
			if assigned == member.ID {
				alreadyAssigned = true
				break
			}
		}
		if alreadyAssigned {
			continue
		}
		candidates = append(candidates, member)
	}

	if newReviewerID == "" {
		if len(candidates) == 0 {
			return domain.PullRequest{}, domain.ErrReviewerNotInTeam
		}
		if uc.rand != nil && len(candidates) > 1 {
			uc.rand.Shuffle(len(candidates), func(i, j int) {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			})
		}
		newReviewerID = candidates[0].ID
	} else {
		valid := false
		for _, candidate := range candidates {
			if candidate.ID == newReviewerID {
				valid = true
				break
			}
		}
		if !valid {
			return domain.PullRequest{}, domain.ErrReviewerNotInTeam
		}
	}

	if err := pr.ReplaceReviewer(oldReviewerID, newReviewerID); err != nil {
		return domain.PullRequest{}, err
	}

	if err := uc.prStorage.UpdatePullRequest(ctx, pr); err != nil {
		return domain.PullRequest{}, err
	}

	return pr, nil
}
