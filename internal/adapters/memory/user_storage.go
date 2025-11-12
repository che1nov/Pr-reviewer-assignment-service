package memory

import (
	"context"
	"sync"
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
