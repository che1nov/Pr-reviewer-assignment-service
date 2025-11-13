package memory

import (
	"context"
	"sync"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type UserStorage struct {
	mu    sync.RWMutex
	users map[string]domain.User
	teams map[string]domain.Team
	prs   map[string]domain.PullRequest
}

func NewUserStorage() *UserStorage {
	return &UserStorage{
		users: make(map[string]domain.User),
		teams: make(map[string]domain.Team),
		prs:   make(map[string]domain.PullRequest),
	}
}

func (s *UserStorage) CreateUser(_ context.Context, user domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[user.ID] = user
	return nil
}

func (s *UserStorage) ListUsers(_ context.Context) ([]domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]domain.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	return users, nil
}

func (s *UserStorage) GetUser(_ context.Context, id string) (domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

func (s *UserStorage) UpdateUser(_ context.Context, user domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[user.ID]; !ok {
		return domain.ErrUserNotFound
	}
	s.users[user.ID] = user

	for name, team := range s.teams {
		updated := false
		for i, member := range team.Users {
			if member.ID == user.ID {
				team.Users[i] = user
				updated = true
			}
		}
		if updated {
			s.teams[name] = team
		}
	}

	return nil
}

func (s *UserStorage) CreateTeam(_ context.Context, team domain.Team) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.teams[team.Name] = team
	return nil
}

func (s *UserStorage) ListTeams(_ context.Context) ([]domain.Team, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	teams := make([]domain.Team, 0, len(s.teams))
	for _, team := range s.teams {
		teams = append(teams, team)
	}

	return teams, nil
}

func (s *UserStorage) GetTeam(_ context.Context, name string) (domain.Team, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	team, ok := s.teams[name]
	if !ok {
		return domain.Team{}, domain.ErrTeamNotFound
	}
	return team, nil
}

func (s *UserStorage) CreatePullRequest(_ context.Context, pr domain.PullRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prs[pr.ID] = pr
	return nil
}

func (s *UserStorage) ListPullRequests(_ context.Context) ([]domain.PullRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prs := make([]domain.PullRequest, 0, len(s.prs))
	for _, pr := range s.prs {
		prs = append(prs, pr)
	}

	return prs, nil
}

func (s *UserStorage) GetPullRequest(_ context.Context, id string) (domain.PullRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pr, ok := s.prs[id]
	if !ok {
		return domain.PullRequest{}, domain.ErrPullRequestNotFound
	}
	return pr, nil
}

func (s *UserStorage) UpdatePullRequest(_ context.Context, pr domain.PullRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.prs[pr.ID]; !ok {
		return domain.ErrPullRequestNotFound
	}
	s.prs[pr.ID] = pr
	return nil
}
