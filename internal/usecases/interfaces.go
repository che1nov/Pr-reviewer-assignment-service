package usecases

import (
	"context"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type UserStorage interface {
	CreateUser(ctx context.Context, id, name string) error
	ListUsers(ctx context.Context) ([]domain.User, error)
}

type TeamStorage interface {
	CreateTeam(ctx context.Context, team domain.Team) error
	ListTeams(ctx context.Context) ([]domain.Team, error)
}

type PullRequestStorage interface {
	CreatePullRequest(ctx context.Context, pr domain.PullRequest) error
	ListPullRequests(ctx context.Context) ([]domain.PullRequest, error)
}
