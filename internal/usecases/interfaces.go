package usecases

import (
	"context"
	"time"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type UserStorage interface {
	CreateUser(ctx context.Context, user domain.User) error
	ListUsers(ctx context.Context) ([]domain.User, error)
	GetUser(ctx context.Context, id string) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) error
}

type TeamStorage interface {
	CreateTeam(ctx context.Context, team domain.Team) error
	ListTeams(ctx context.Context) ([]domain.Team, error)
	GetTeam(ctx context.Context, name string) (domain.Team, error)
}

type PullRequestStorage interface {
	CreatePullRequest(ctx context.Context, pr domain.PullRequest) error
	ListPullRequests(ctx context.Context) ([]domain.PullRequest, error)
	GetPullRequest(ctx context.Context, id string) (domain.PullRequest, error)
	UpdatePullRequest(ctx context.Context, pr domain.PullRequest) error
	ListPullRequestsByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error)
}

type ClockAdapter interface {
	Now() time.Time
}

type RandomAdapter interface {
	Shuffle(n int, swap func(i, j int))
}
