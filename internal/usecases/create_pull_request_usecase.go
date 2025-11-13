package usecases

import (
	"context"
	"math/rand"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type CreatePullRequestUseCase struct {
	prStorage   PullRequestStorage
	teamStorage TeamStorage
	rand        *rand.Rand
}

func NewCreatePullRequestUseCase(prStorage PullRequestStorage, teamStorage TeamStorage, rng *rand.Rand) *CreatePullRequestUseCase {
	return &CreatePullRequestUseCase{
		prStorage:   prStorage,
		teamStorage: teamStorage,
		rand:        rng,
	}
}

func (uc *CreatePullRequestUseCase) Execute(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error) {
	team, err := uc.teamStorage.GetTeam(ctx, pr.TeamName)
	if err != nil {
		return domain.PullRequest{}, err
	}

	reviewers := team.ActiveReviewersExcluding(pr.AuthorID)
	if len(reviewers) > 1 && uc.rand != nil {
		uc.rand.Shuffle(len(reviewers), func(i, j int) {
			reviewers[i], reviewers[j] = reviewers[j], reviewers[i]
		})
	}
	if len(reviewers) > 2 {
		reviewers = reviewers[:2]
	}

	reviewerIDs := make([]string, 0, len(reviewers))
	for _, reviewer := range reviewers {
		reviewerIDs = append(reviewerIDs, reviewer.ID)
	}

	pr.AssignReviewers(reviewerIDs)

	if err := uc.prStorage.CreatePullRequest(ctx, pr); err != nil {
		return domain.PullRequest{}, err
	}

	return pr, nil
}
