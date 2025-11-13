package memory

import (
	"context"
	"sync"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type UserStorage struct {
	mu    sync.RWMutex
	users map[string]string
	teams map[string]domain.Team
}

func NewUserStorage() *UserStorage {
	return &UserStorage{
		users: make(map[string]string),
		teams: make(map[string]domain.Team),
	}
}

func (s *UserStorage) CreateUser(_ context.Context, id, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[id] = name
	return nil
}

func (s *UserStorage) ListUsers(_ context.Context) ([]domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]domain.User, 0, len(s.users))
	for id, name := range s.users {
		users = append(users, domain.User{ID: id, Name: name})
	}

	return users, nil
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
