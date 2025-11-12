package memory

import (
	"context"
	"sync"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type UserStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewUserStorage() *UserStorage {
	return &UserStorage{
		data: make(map[string]string),
	}
}

func (s *UserStorage) CreateUser(_ context.Context, id, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = name
	return nil
}

func (s *UserStorage) ListUsers(_ context.Context) ([]domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]domain.User, 0, len(s.data))
	for id, name := range s.data {
		users = append(users, domain.User{ID: id, Name: name})
	}

	return users, nil
}
