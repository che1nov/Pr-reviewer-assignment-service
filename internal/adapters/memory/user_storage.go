package memory

import (
	"context"
	"log/slog"
	"sync"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/domain"
)

type UserStorage struct {
	mu    sync.RWMutex
	users map[string]domain.User
	teams map[string]domain.Team
	prs   map[string]domain.PullRequest
	log   *slog.Logger
}

func NewUserStorage(log *slog.Logger) *UserStorage {
	return &UserStorage{
		users: make(map[string]domain.User),
		teams: make(map[string]domain.Team),
		prs:   make(map[string]domain.PullRequest),
		log:   log,
	}
}

// CreateUser сохраняет пользователя.
func (s *UserStorage) CreateUser(_ context.Context, user domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[user.ID] = user
	s.log.Debug("создан пользователь", "user_id", user.ID, "team_name", user.TeamName)
	return nil
}

// ListUsers возвращает всех пользователей.
func (s *UserStorage) ListUsers(_ context.Context) ([]domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]domain.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	return users, nil
}

// GetUser возвращает пользователя.
func (s *UserStorage) GetUser(_ context.Context, id string) (domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

// UpdateUser обновляет пользователя.
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

	s.log.Debug("обновлён пользователь", "user_id", user.ID, "team_name", user.TeamName, "is_active", user.IsActive)
	return nil
}

// CreateTeam сохраняет команду.
func (s *UserStorage) CreateTeam(_ context.Context, team domain.Team) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.teams[team.Name]; exists {
		return domain.ErrTeamExists
	}
	s.teams[team.Name] = team
	for _, member := range team.Users {
		if _, ok := s.users[member.ID]; !ok {
			s.users[member.ID] = member
		}
	}
	s.log.Debug("создана команда", "team_name", team.Name, "members", len(team.Users))
	return nil
}

// ListTeams возвращает все команды.
func (s *UserStorage) ListTeams(_ context.Context) ([]domain.Team, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	teams := make([]domain.Team, 0, len(s.teams))
	for _, team := range s.teams {
		teams = append(teams, team)
	}

	return teams, nil
}

// GetTeam возвращает команду.
func (s *UserStorage) GetTeam(_ context.Context, name string) (domain.Team, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	team, ok := s.teams[name]
	if !ok {
		return domain.Team{}, domain.ErrTeamNotFound
	}
	return team, nil
}

// CreatePullRequest сохраняет pull request.
func (s *UserStorage) CreatePullRequest(_ context.Context, pr domain.PullRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.prs[pr.ID]; exists {
		return domain.ErrPullRequestExists
	}
	s.prs[pr.ID] = pr
	s.log.Debug("создан pull request", "pr_id", pr.ID, "author_id", pr.AuthorID, "reviewers", pr.Reviewers)
	return nil
}

// ListPullRequests возвращает pull request.
func (s *UserStorage) ListPullRequests(_ context.Context) ([]domain.PullRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prs := make([]domain.PullRequest, 0, len(s.prs))
	for _, pr := range s.prs {
		prs = append(prs, pr)
	}

	return prs, nil
}

// GetPullRequest возвращает pull request.
func (s *UserStorage) GetPullRequest(_ context.Context, id string) (domain.PullRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pr, ok := s.prs[id]
	if !ok {
		return domain.PullRequest{}, domain.ErrPullRequestNotFound
	}
	return pr, nil
}

// UpdatePullRequest обновляет pull request.
func (s *UserStorage) UpdatePullRequest(_ context.Context, pr domain.PullRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.prs[pr.ID]; !ok {
		return domain.ErrPullRequestNotFound
	}
	s.prs[pr.ID] = pr
	s.log.Debug("обновлён pull request", "pr_id", pr.ID, "status", pr.Status, "reviewers", pr.Reviewers)
	return nil
}

// ListPullRequestsByReviewer возвращает pull request по ревьюеру.
func (s *UserStorage) ListPullRequestsByReviewer(_ context.Context, reviewerID string) ([]domain.PullRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.PullRequest, 0)
	for _, pr := range s.prs {
		for _, reviewer := range pr.Reviewers {
			if reviewer == reviewerID {
				result = append(result, pr)
				break
			}
		}
	}

	return result, nil
}
