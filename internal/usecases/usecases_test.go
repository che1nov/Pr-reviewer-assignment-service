package usecases

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/che1nov/Pr-reviewer-assignment-service/internal/domain"
	"github.com/che1nov/Pr-reviewer-assignment-service/internal/dto"
)

func TestCreateTeamUseCase_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	errGetTeam := errors.New("team storage failure")
	errGetUser := errors.New("user storage failure")
	errCreateUser := errors.New("create user failure")
	errUpdateUser := errors.New("update user failure")
	errCreateTeam := errors.New("create team failure")

	tests := []struct {
		name          string
		initialTeams  []domain.Team
		initialUsers  []domain.User
		input         domain.Team
		wantErr       error
		configure     func(teamStorage *fakeTeamStorage, userStorage *fakeUserStorage)
		verifySuccess func(t *testing.T, team domain.Team, users *fakeUserStorage)
	}{
		{
			name:         "success with existing member update",
			initialTeams: nil,
			initialUsers: []domain.User{domain.NewUser("u1", "Legacy", "legacy", false)},
			input: domain.Team{
				Name: "backend",
				Users: []domain.User{
					domain.NewUser("u1", "Alice", "", true),
					domain.NewUser("u2", "Bob", "", true),
				},
			},
			verifySuccess: func(t *testing.T, team domain.Team, users *fakeUserStorage) {
				t.Helper()
				if team.Name != "backend" {
					t.Fatalf("expected team backend, got %q", team.Name)
				}
				if len(team.Users) != 2 {
					t.Fatalf("expected 2 users, got %d", len(team.Users))
				}
				if users.users["u1"].TeamName != "backend" || !users.users["u1"].IsActive {
					t.Fatalf("expected legacy user moved to backend and active, got %#v", users.users["u1"])
				}
				if users.users["u2"].TeamName != "backend" {
					t.Fatalf("expected new user team backend, got %q", users.users["u2"].TeamName)
				}
			},
		},
		{
			name:         "team already exists",
			initialTeams: []domain.Team{domain.NewTeam("backend", nil)},
			initialUsers: nil,
			input:        domain.Team{Name: "backend"},
			wantErr:      domain.ErrTeamExists,
		},
		{
			name:  "get team returns unexpected error",
			input: domain.Team{Name: "backend"},
			configure: func(teamStorage *fakeTeamStorage, _ *fakeUserStorage) {
				teamStorage.getErr = errGetTeam
				teamStorage.getErrName = "backend"
			},
			wantErr: errGetTeam,
		},
		{
			name:  "get user returns unexpected error",
			input: domain.Team{Name: "backend", Users: []domain.User{domain.NewUser("u1", "Alice", "", true)}},
			configure: func(_ *fakeTeamStorage, userStorage *fakeUserStorage) {
				userStorage.getErr = errGetUser
				userStorage.getErrID = "u1"
			},
			wantErr: errGetUser,
		},
		{
			name:  "create user returns error",
			input: domain.Team{Name: "backend", Users: []domain.User{domain.NewUser("u1", "Alice", "", true)}},
			configure: func(_ *fakeTeamStorage, userStorage *fakeUserStorage) {
				userStorage.createErr = errCreateUser
				userStorage.createErrID = "u1"
			},
			wantErr: errCreateUser,
		},
		{
			name:         "update user returns error",
			initialUsers: []domain.User{domain.NewUser("u1", "Legacy", "legacy", false)},
			input: domain.Team{
				Name: "backend",
				Users: []domain.User{
					domain.NewUser("u1", "Alice", "", true),
				},
			},
			configure: func(_ *fakeTeamStorage, userStorage *fakeUserStorage) {
				userStorage.updateErr = errUpdateUser
				userStorage.updateErrID = "u1"
			},
			wantErr: errUpdateUser,
		},
		{
			name:  "create team returns error",
			input: domain.Team{Name: "backend"},
			configure: func(teamStorage *fakeTeamStorage, _ *fakeUserStorage) {
				teamStorage.createErr = errCreateTeam
				teamStorage.createErrName = "backend"
			},
			wantErr: errCreateTeam,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userStorage := newFakeUserStorage(tt.initialUsers...)
			teamStorage := newFakeTeamStorage(tt.initialTeams...)
			uc := NewCreateTeamUseCase(teamStorage, userStorage, testLogger())

			if tt.configure != nil {
				tt.configure(teamStorage, userStorage)
			}

			result, err := uc.Create(ctx, tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}

			if tt.wantErr == nil && tt.verifySuccess != nil {
				tt.verifySuccess(t, result, userStorage)
			}
		})
	}
}

func TestGetTeamUseCase_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	errStorage := errors.New("storage failure")

	tests := []struct {
		name         string
		initialTeams []domain.Team
		requestName  string
		configure    func(storage *fakeTeamStorage)
		wantErr      error
		verify       func(t *testing.T, team domain.Team)
	}{
		{
			name:         "team found",
			initialTeams: []domain.Team{domain.NewTeam("backend", nil)},
			requestName:  "backend",
			verify: func(t *testing.T, team domain.Team) {
				t.Helper()
				if team.Name != "backend" {
					t.Fatalf("expected team backend, got %s", team.Name)
				}
			},
		},
		{
			name: "team found with members",
			initialTeams: []domain.Team{
				domain.NewTeam("backend", []domain.User{
					domain.NewUser("u1", "Alice", "backend", true),
					domain.NewUser("u2", "Bob", "backend", true),
				}),
			},
			requestName: "backend",
			verify: func(t *testing.T, team domain.Team) {
				t.Helper()
				if team.Name != "backend" {
					t.Fatalf("expected team backend, got %s", team.Name)
				}
				if len(team.Users) != 2 {
					t.Fatalf("expected 2 members, got %d", len(team.Users))
				}
			},
		},
		{
			name:        "team not found",
			requestName: "unknown",
			wantErr:     domain.ErrTeamNotFound,
		},
		{
			name:        "storage error",
			requestName: "backend",
			configure: func(storage *fakeTeamStorage) {
				storage.getErr = errStorage
				storage.getErrName = "backend"
			},
			wantErr: errStorage,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			teamStorage := newFakeTeamStorage(tt.initialTeams...)
			if tt.configure != nil {
				tt.configure(teamStorage)
			}
			uc := NewGetTeamUseCase(teamStorage, testLogger())

			team, err := uc.Get(ctx, tt.requestName)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && tt.verify != nil {
				tt.verify(t, team)
			}
		})
	}
}

func TestCreatePullRequestUseCase_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baseAuthor := domain.NewUser("author", "Alice", "backend", true)
	baseTeam := domain.NewTeam("backend", []domain.User{baseAuthor})
	clock := fakeClock{now: time.Unix(42, 0)}
	random := &fakeRandom{}

	tests := []struct {
		name       string
		users      []domain.User
		team       domain.Team
		initialPRs []domain.PullRequest
		wantErr    error
		verify     func(t *testing.T, pr domain.PullRequest, storage *fakePullRequestStorage)
	}{
		{
			name: "success assigns reviewers",
			users: []domain.User{
				baseAuthor,
				domain.NewUser("r1", "Bob", "backend", true),
				domain.NewUser("r2", "Charlie", "backend", true),
				domain.NewUser("inactive", "Dave", "backend", false),
			},
			team: domain.NewTeam("backend", []domain.User{
				baseAuthor,
				domain.NewUser("r1", "Bob", "backend", true),
				domain.NewUser("r2", "Charlie", "backend", true),
				domain.NewUser("inactive", "Dave", "backend", false),
			}),
			verify: func(t *testing.T, pr domain.PullRequest, storage *fakePullRequestStorage) {
				t.Helper()
				if pr.Status != "OPEN" {
					t.Fatalf("expected status OPEN, got %q", pr.Status)
				}
				if len(pr.Reviewers) != 2 {
					t.Fatalf("expected two reviewers, got %v", pr.Reviewers)
				}
				for _, reviewer := range pr.Reviewers {
					if reviewer == "author" || reviewer == "inactive" {
						t.Fatalf("unexpected reviewer %q", reviewer)
					}
				}
				if _, err := storage.GetPullRequest(context.Background(), "pr-1"); err != nil {
					t.Fatalf("expected pull request persisted: %v", err)
				}
			},
		},
		{
			name:       "pull request exists",
			users:      []domain.User{baseAuthor},
			team:       baseTeam,
			initialPRs: []domain.PullRequest{domain.NewPullRequest("pr-1", "Existing", "author", "backend", time.Now())},
			wantErr:    domain.ErrPullRequestExists,
		},
		{
			name:    "author not found",
			users:   nil,
			team:    baseTeam,
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:    "team not found",
			users:   []domain.User{baseAuthor},
			team:    domain.Team{},
			wantErr: domain.ErrTeamNotFound,
		},
		{
			name:  "no reviewer candidates assigns none",
			users: []domain.User{baseAuthor},
			team:  domain.NewTeam("backend", []domain.User{baseAuthor}),
			verify: func(t *testing.T, pr domain.PullRequest, storage *fakePullRequestStorage) {
				t.Helper()
				if len(pr.Reviewers) != 0 {
					t.Fatalf("expected no reviewers, got %v", pr.Reviewers)
				}
				if _, err := storage.GetPullRequest(context.Background(), "pr-1"); err != nil {
					t.Fatalf("expected pull request persisted: %v", err)
				}
			},
		},
		{
			name: "only one active candidate assigns one reviewer",
			users: []domain.User{
				baseAuthor,
				domain.NewUser("r1", "Bob", "backend", true),
			},
			team: domain.NewTeam("backend", []domain.User{
				baseAuthor,
				domain.NewUser("r1", "Bob", "backend", true),
			}),
			verify: func(t *testing.T, pr domain.PullRequest, storage *fakePullRequestStorage) {
				t.Helper()
				if len(pr.Reviewers) != 1 {
					t.Fatalf("expected one reviewer, got %v", pr.Reviewers)
				}
				if pr.Reviewers[0] != "r1" {
					t.Fatalf("expected reviewer r1, got %s", pr.Reviewers[0])
				}
			},
		},
		{
			name: "all team members inactive except author",
			users: []domain.User{
				baseAuthor,
				domain.NewUser("r1", "Bob", "backend", false),
				domain.NewUser("r2", "Charlie", "backend", false),
			},
			team: domain.NewTeam("backend", []domain.User{
				baseAuthor,
				domain.NewUser("r1", "Bob", "backend", false),
				domain.NewUser("r2", "Charlie", "backend", false),
			}),
			verify: func(t *testing.T, pr domain.PullRequest, storage *fakePullRequestStorage) {
				t.Helper()
				if len(pr.Reviewers) != 0 {
					t.Fatalf("expected no reviewers (all inactive), got %v", pr.Reviewers)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prStorage := newFakePullRequestStorage(tt.initialPRs...)
			userStorage := newFakeUserStorage(tt.users...)
			teamStorage := newFakeTeamStorage()
			if tt.team.Name != "" {
				teamStorage = newFakeTeamStorage(tt.team)
			}

			uc := NewCreatePullRequestUseCase(prStorage, teamStorage, userStorage, clock, random, testLogger())
			pr, err := uc.Create(ctx, "pr-1", "Feature", "author")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && tt.verify != nil {
				tt.verify(t, pr, prStorage)
			}
		})
	}
}

func TestMergePullRequestUseCase_Merge(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	clock := fakeClock{now: time.Unix(99, 0)}

	tests := []struct {
		name       string
		initialPRs []domain.PullRequest
		id         string
		wantErr    error
		verify     func(t *testing.T, pr domain.PullRequest)
	}{
		{
			name:       "merge success",
			initialPRs: []domain.PullRequest{domain.NewPullRequest("pr-1", "Feature", "author", "backend", time.Now())},
			id:         "pr-1",
			verify: func(t *testing.T, pr domain.PullRequest) {
				t.Helper()
				if pr.Status != "MERGED" || pr.MergedAt == nil {
					t.Fatalf("expected merged pull request, got %#v", pr)
				}
				expectedMergedAt := time.Unix(99, 0)
				if !pr.MergedAt.Equal(expectedMergedAt) {
					t.Fatalf("expected mergedAt %v, got %v", expectedMergedAt, *pr.MergedAt)
				}
			},
		},
		{
			name: "merge already merged PR is idempotent",
			initialPRs: []domain.PullRequest{func() domain.PullRequest {
				pr := domain.NewPullRequest("pr-1", "Feature", "author", "backend", time.Now())
				originalMergeTime := time.Unix(50, 0)
				pr.Status = "MERGED"
				pr.MergedAt = &originalMergeTime
				return pr
			}()},
			id: "pr-1",
			verify: func(t *testing.T, pr domain.PullRequest) {
				t.Helper()
				if pr.Status != "MERGED" {
					t.Fatalf("expected status MERGED, got %s", pr.Status)
				}
				if pr.MergedAt == nil {
					t.Fatalf("expected MergedAt to be set")
				}
				// Проверяем, что время не изменилось (идемпотентность)
				originalTime := time.Unix(50, 0)
				if !pr.MergedAt.Equal(originalTime) {
					t.Fatalf("expected original merge time %v to be preserved, got %v", originalTime, *pr.MergedAt)
				}
			},
		},
		{
			name:    "pull request not found",
			id:      "unknown",
			wantErr: domain.ErrPullRequestNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prStorage := newFakePullRequestStorage(tt.initialPRs...)
			uc := NewMergePullRequestUseCase(prStorage, clock, testLogger())

			pr, err := uc.Merge(ctx, tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && tt.verify != nil {
				tt.verify(t, pr)
			}
		})
	}
}

func TestReassignReviewerUseCase_Reassign(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	author := domain.NewUser("author", "Alice", "backend", true)
	oldReviewer := domain.NewUser("old", "Bob", "backend", true)

	errUserFetch := errors.New("user fetch failure")
	errUpdatePR := errors.New("update pull request failure")

	tests := []struct {
		name       string
		prStore    *fakePullRequestStorage
		team       *fakeTeamStorage
		users      *fakeUserStorage
		desiredNew *string
		wantErr    error
		verify     func(t *testing.T, pr domain.PullRequest, replacedBy string)
	}{
		{
			name: "success replaces reviewer",
			prStore: newFakePullRequestStorage(func() domain.PullRequest {
				pr := domain.NewPullRequest("pr-1", "Feature", "author", "backend", time.Now())
				pr.AssignReviewers([]string{"old", "busy"})
				return pr
			}()),
			team: newFakeTeamStorage(domain.NewTeam("backend", []domain.User{
				author,
				oldReviewer,
				domain.NewUser("busy", "Busy", "backend", true),
				domain.NewUser("candidate", "Charlie", "backend", true),
			})),
			users: newFakeUserStorage(oldReviewer, domain.NewUser("candidate", "Charlie", "backend", true)),
			verify: func(t *testing.T, pr domain.PullRequest, replacedBy string) {
				t.Helper()
				if replacedBy == "" || replacedBy == "old" {
					t.Fatalf("unexpected replaced reviewer %q", replacedBy)
				}
				if len(pr.Reviewers) == 0 || pr.Reviewers[0] == "old" {
					t.Fatalf("expected reviewer replaced, got %v", pr.Reviewers)
				}
			},
		},
		{
			name:    "pull request not found",
			prStore: newFakePullRequestStorage(),
			team:    newFakeTeamStorage(domain.NewTeam("backend", []domain.User{author, oldReviewer})),
			users:   newFakeUserStorage(oldReviewer),
			wantErr: domain.ErrPullRequestNotFound,
		},
		{
			name: "no candidates",
			prStore: newFakePullRequestStorage(func() domain.PullRequest {
				pr := domain.NewPullRequest("pr-1", "Feature", "author", "backend", time.Now())
				pr.AssignReviewers([]string{"old"})
				return pr
			}()),
			team:    newFakeTeamStorage(domain.NewTeam("backend", []domain.User{author, oldReviewer})),
			users:   newFakeUserStorage(oldReviewer),
			wantErr: domain.ErrNoReviewerCandidates,
		},
		{
			name: "new reviewer not found",
			prStore: newFakePullRequestStorage(func() domain.PullRequest {
				pr := domain.NewPullRequest("pr-1", "Feature", "author", "backend", time.Now())
				pr.AssignReviewers([]string{"old"})
				return pr
			}()),
			team: newFakeTeamStorage(domain.NewTeam("backend", []domain.User{
				author,
				oldReviewer,
				domain.NewUser("candidate", "Charlie", "backend", true),
			})),
			users:   newFakeUserStorage(oldReviewer),
			wantErr: domain.ErrUserNotFound,
		},
		{
			name: "desired reviewer not a candidate",
			prStore: newFakePullRequestStorage(func() domain.PullRequest {
				pr := domain.NewPullRequest("pr-1", "Feature", "author", "backend", time.Now())
				pr.AssignReviewers([]string{"old"})
				return pr
			}()),
			team: newFakeTeamStorage(domain.NewTeam("backend", []domain.User{
				author,
				oldReviewer,
				domain.NewUser("inactive", "Charlie", "backend", false),
			})),
			users:      newFakeUserStorage(oldReviewer),
			desiredNew: stringPtr("inactive"),
			wantErr:    domain.ErrNoReviewerCandidates,
		},
		{
			name: "desired reviewer success",
			prStore: newFakePullRequestStorage(func() domain.PullRequest {
				pr := domain.NewPullRequest("pr-1", "Feature", "author", "backend", time.Now())
				pr.AssignReviewers([]string{"old", "busy"})
				return pr
			}()),
			team: newFakeTeamStorage(domain.NewTeam("backend", []domain.User{
				author,
				oldReviewer,
				domain.NewUser("busy", "Busy", "backend", true),
				domain.NewUser("candidate", "Charlie", "backend", true),
			})),
			users:      newFakeUserStorage(oldReviewer, domain.NewUser("candidate", "Charlie", "backend", true)),
			desiredNew: stringPtr("candidate"),
			verify: func(t *testing.T, pr domain.PullRequest, replacedBy string) {
				t.Helper()
				if replacedBy != "candidate" {
					t.Fatalf("expected candidate replacement, got %q", replacedBy)
				}
				if contains(pr.Reviewers, "old") {
					t.Fatalf("expected old reviewer removed, got %v", pr.Reviewers)
				}
			},
		},
		{
			name: "new reviewer load failure",
			prStore: newFakePullRequestStorage(func() domain.PullRequest {
				pr := domain.NewPullRequest("pr-1", "Feature", "author", "backend", time.Now())
				pr.AssignReviewers([]string{"old"})
				return pr
			}()),
			team: newFakeTeamStorage(domain.NewTeam("backend", []domain.User{
				author,
				oldReviewer,
				domain.NewUser("candidate", "Charlie", "backend", true),
			})),
			users: func() *fakeUserStorage {
				store := newFakeUserStorage(oldReviewer, domain.NewUser("candidate", "Charlie", "backend", true))
				store.getErr = errUserFetch
				store.getErrID = "candidate"
				return store
			}(),
			wantErr: errUserFetch,
		},
		{
			name: "update pull request failure",
			prStore: func() *fakePullRequestStorage {
				store := newFakePullRequestStorage(func() domain.PullRequest {
					pr := domain.NewPullRequest("pr-1", "Feature", "author", "backend", time.Now())
					pr.AssignReviewers([]string{"old"})
					return pr
				}())
				store.updateErr = errUpdatePR
				store.updateErrID = "pr-1"
				return store
			}(),
			team: newFakeTeamStorage(domain.NewTeam("backend", []domain.User{
				author,
				oldReviewer,
				domain.NewUser("candidate", "Charlie", "backend", true),
			})),
			users:   newFakeUserStorage(oldReviewer, domain.NewUser("candidate", "Charlie", "backend", true)),
			wantErr: errUpdatePR,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := NewReassignReviewerUseCase(tt.prStore, tt.team, tt.users, &fakeRandom{}, testLogger())
			pr, replacedBy, err := uc.Reassign(ctx, "pr-1", "old", tt.desiredNew)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && tt.verify != nil {
				tt.verify(t, pr, replacedBy)
			}
		})
	}
}

func TestSetUserActiveUseCase_SetActive(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	errUpdate := errors.New("update failure")

	tests := []struct {
		name      string
		users     []domain.User
		userID    string
		active    bool
		configure func(storage *fakeUserStorage)
		wantErr   error
		verify    func(t *testing.T, storage *fakeUserStorage, result domain.User)
	}{
		{
			name:   "activate user",
			users:  []domain.User{domain.NewUser("u1", "Alice", "backend", false)},
			userID: "u1",
			active: true,
			verify: func(t *testing.T, storage *fakeUserStorage, result domain.User) {
				t.Helper()
				if !storage.users["u1"].IsActive {
					t.Fatalf("expected user active in storage")
				}
				if !result.IsActive {
					t.Fatalf("expected result user active")
				}
			},
		},
		{
			name:   "deactivate user",
			users:  []domain.User{domain.NewUser("u1", "Alice", "backend", true)},
			userID: "u1",
			active: false,
			verify: func(t *testing.T, storage *fakeUserStorage, result domain.User) {
				t.Helper()
				if storage.users["u1"].IsActive {
					t.Fatalf("expected user inactive in storage")
				}
				if result.IsActive {
					t.Fatalf("expected result user inactive")
				}
			},
		},
		{
			name:    "user not found",
			userID:  "missing",
			active:  true,
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:   "update user returns error",
			users:  []domain.User{domain.NewUser("u1", "Alice", "backend", false)},
			userID: "u1",
			active: true,
			configure: func(storage *fakeUserStorage) {
				storage.updateErr = errUpdate
				storage.updateErrID = "u1"
			},
			wantErr: errUpdate,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userStorage := newFakeUserStorage(tt.users...)
			if tt.configure != nil {
				tt.configure(userStorage)
			}
			uc := NewSetUserActiveUseCase(userStorage, testLogger())

			result, err := uc.SetActive(ctx, tt.userID, tt.active)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && tt.verify != nil {
				tt.verify(t, userStorage, result)
			}
		})
	}
}

func TestGetReviewerPullRequestsUseCase_ListByReviewer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		initialPRs  []domain.PullRequest
		reviewerID  string
		setupErr    error
		wantErr     error
		verifyCount int
	}{
		{
			name: "success filters reviewer",
			initialPRs: []domain.PullRequest{
				func() domain.PullRequest {
					pr := domain.NewPullRequest("pr-1", "A", "author", "backend", time.Now())
					pr.AssignReviewers([]string{"r1"})
					return pr
				}(),
				func() domain.PullRequest {
					pr := domain.NewPullRequest("pr-2", "B", "author", "backend", time.Now())
					pr.AssignReviewers([]string{"r2"})
					return pr
				}(),
			},
			reviewerID:  "r1",
			verifyCount: 1,
		},
		{
			name:       "storage failure",
			reviewerID: "r1",
			setupErr:   errors.New("storage failure"),
			wantErr:    errors.New("storage failure"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prStorage := newFakePullRequestStorage(tt.initialPRs...)
			prStorage.listByReviewerErr = tt.setupErr

			uc := NewGetReviewerPullRequestsUseCase(prStorage, testLogger())
			result, err := uc.ListByReviewer(ctx, tt.reviewerID)

			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != tt.verifyCount {
				t.Fatalf("expected %d pull requests, got %d", tt.verifyCount, len(result))
			}
		})
	}
}

func TestListPullRequestsUseCase_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	errList := errors.New("list failure")

	tests := []struct {
		name        string
		initialPRs  []domain.PullRequest
		expectedLen int
		wantErr     error
	}{
		{
			name: "returns all pull requests",
			initialPRs: []domain.PullRequest{
				domain.NewPullRequest("pr-1", "A", "author", "backend", time.Now()),
				domain.NewPullRequest("pr-2", "B", "author", "backend", time.Now()),
			},
			expectedLen: 2,
		},
		{
			name:    "storage error",
			wantErr: errList,
		},
		{
			name:        "empty list",
			initialPRs:  []domain.PullRequest{},
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prStorage := newFakePullRequestStorage(tt.initialPRs...)
			if tt.wantErr != nil {
				prStorage.listErr = tt.wantErr
			}
			uc := NewListPullRequestsUseCase(prStorage, testLogger())

			result, err := uc.List(ctx)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil {
				return
			}
			if len(result) != tt.expectedLen {
				t.Fatalf("expected %d pull requests, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

func TestGetStatsUseCase_GetStats(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	errPRList := errors.New("pr list failure")
	errUserList := errors.New("user list failure")

	tests := []struct {
		name       string
		initialPRs []domain.PullRequest
		users      []domain.User
		prErr      error
		userErr    error
		wantErr    error
		verify     func(t *testing.T, stats dto.StatsResponse)
	}{
		{
			name: "success with multiple PRs and reviewers",
			initialPRs: []domain.PullRequest{
				func() domain.PullRequest {
					pr := domain.NewPullRequest("pr-1", "Feature A", "u1", "backend", time.Now())
					pr.AssignReviewers([]string{"u2", "u3"})
					return pr
				}(),
				func() domain.PullRequest {
					pr := domain.NewPullRequest("pr-2", "Feature B", "u2", "backend", time.Now())
					pr.AssignReviewers([]string{"u3"})
					merged := time.Now()
					pr.MergedAt = &merged
					pr.Status = "MERGED"
					return pr
				}(),
				func() domain.PullRequest {
					pr := domain.NewPullRequest("pr-3", "Feature C", "u3", "backend", time.Now())
					pr.AssignReviewers([]string{"u1"})
					return pr
				}(),
			},
			users: []domain.User{
				domain.NewUser("u1", "Alice", "backend", true),
				domain.NewUser("u2", "Bob", "backend", true),
				domain.NewUser("u3", "Charlie", "backend", true),
			},
			verify: func(t *testing.T, stats dto.StatsResponse) {
				t.Helper()
				if stats.PRStats.TotalPRs != 3 {
					t.Fatalf("expected 3 total PRs, got %d", stats.PRStats.TotalPRs)
				}
				if stats.PRStats.OpenPRs != 2 {
					t.Fatalf("expected 2 open PRs, got %d", stats.PRStats.OpenPRs)
				}
				if stats.PRStats.MergedPRs != 1 {
					t.Fatalf("expected 1 merged PR, got %d", stats.PRStats.MergedPRs)
				}
				if stats.PRStats.ReviewersCount != 3 {
					t.Fatalf("expected 3 unique reviewers, got %d", stats.PRStats.ReviewersCount)
				}
				if len(stats.UserStats) != 3 {
					t.Fatalf("expected 3 users with assignments, got %d", len(stats.UserStats))
				}
				// Проверяем, что у каждого пользователя есть назначения
				for _, userStat := range stats.UserStats {
					if userStat.AssignedPRs == 0 {
						t.Fatalf("expected all users to have assignments, but %s has 0", userStat.UserID)
					}
				}
			},
		},
		{
			name:       "empty PRs and users",
			initialPRs: []domain.PullRequest{},
			users:      []domain.User{},
			verify: func(t *testing.T, stats dto.StatsResponse) {
				t.Helper()
				if stats.PRStats.TotalPRs != 0 {
					t.Fatalf("expected 0 total PRs, got %d", stats.PRStats.TotalPRs)
				}
				if len(stats.UserStats) != 0 {
					t.Fatalf("expected 0 user stats, got %d", len(stats.UserStats))
				}
			},
		},
		{
			name: "users without assignments excluded",
			initialPRs: []domain.PullRequest{
				func() domain.PullRequest {
					pr := domain.NewPullRequest("pr-1", "Feature", "u1", "backend", time.Now())
					pr.AssignReviewers([]string{"u2"})
					return pr
				}(),
			},
			users: []domain.User{
				domain.NewUser("u1", "Alice", "backend", true),
				domain.NewUser("u2", "Bob", "backend", true),
				domain.NewUser("u3", "Charlie", "backend", true),
			},
			verify: func(t *testing.T, stats dto.StatsResponse) {
				t.Helper()
				// Только u2 имеет назначения
				if len(stats.UserStats) != 1 {
					t.Fatalf("expected 1 user with assignments, got %d", len(stats.UserStats))
				}
				if stats.UserStats[0].UserID != "u2" {
					t.Fatalf("expected u2 to have assignments, got %s", stats.UserStats[0].UserID)
				}
			},
		},
		{
			name: "PR without reviewers",
			initialPRs: []domain.PullRequest{
				domain.NewPullRequest("pr-1", "Feature", "u1", "backend", time.Now()),
			},
			users: []domain.User{
				domain.NewUser("u1", "Alice", "backend", true),
			},
			verify: func(t *testing.T, stats dto.StatsResponse) {
				t.Helper()
				if stats.PRStats.TotalPRs != 1 {
					t.Fatalf("expected 1 total PR, got %d", stats.PRStats.TotalPRs)
				}
				if stats.PRStats.ReviewersCount != 0 {
					t.Fatalf("expected 0 reviewers, got %d", stats.PRStats.ReviewersCount)
				}
				if len(stats.UserStats) != 0 {
					t.Fatalf("expected 0 user stats, got %d", len(stats.UserStats))
				}
			},
		},
		{
			name:    "PR list error",
			prErr:   errPRList,
			wantErr: errPRList,
		},
		{
			name: "user list error",
			initialPRs: []domain.PullRequest{
				domain.NewPullRequest("pr-1", "Feature", "u1", "backend", time.Now()),
			},
			userErr: errUserList,
			wantErr: errUserList,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prStorage := newFakePullRequestStorage(tt.initialPRs...)
			if tt.prErr != nil {
				prStorage.listErr = tt.prErr
			}

			userStorage := newFakeUserStorage(tt.users...)
			if tt.userErr != nil {
				userStorage.listErr = tt.userErr
			}

			uc := NewGetStatsUseCase(prStorage, userStorage, testLogger())

			stats, err := uc.GetStats(ctx)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.verify != nil {
				tt.verify(t, stats)
			}
		})
	}
}

func TestDeactivateTeamUsersUseCase_DeactivateTeamUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name     string
		teamName string
		users    []domain.User
		teams    []domain.Team
		prs      []domain.PullRequest
		wantErr  error
		verify   func(t *testing.T, result DeactivateResult, userStorage *fakeUserStorage, prStorage *fakePullRequestStorage)
	}{
		{
			name:     "success deactivates all team members and reassigns PRs",
			teamName: "backend",
			users: []domain.User{
				domain.NewUser("u1", "Alice", "backend", true),
				domain.NewUser("u2", "Bob", "backend", true),
				domain.NewUser("u3", "Charlie", "backend", true),
			},
			teams: []domain.Team{
				domain.NewTeam("backend", []domain.User{
					domain.NewUser("u1", "Alice", "backend", true),
					domain.NewUser("u2", "Bob", "backend", true),
					domain.NewUser("u3", "Charlie", "backend", true),
				}),
			},
			prs: []domain.PullRequest{
				func() domain.PullRequest {
					pr := domain.NewPullRequest("pr-1", "Fix bug", "u1", "backend", time.Now())
					pr.AssignReviewers([]string{"u2", "u3"})
					return pr
				}(),
				func() domain.PullRequest {
					pr := domain.NewPullRequest("pr-2", "Add feature", "u2", "backend", time.Now())
					pr.AssignReviewers([]string{"u1"})
					return pr
				}(),
			},
			verify: func(t *testing.T, result DeactivateResult, userStorage *fakeUserStorage, prStorage *fakePullRequestStorage) {
				t.Helper()
				if result.DeactivatedCount != 3 {
					t.Fatalf("expected 3 deactivations, got %d", result.DeactivatedCount)
				}
				for _, userID := range []string{"u1", "u2", "u3"} {
					if userStorage.users[userID].IsActive {
						t.Fatalf("expected user %s to be inactive", userID)
					}
				}
			},
		},
		{
			name:     "empty team returns zero counts",
			teamName: "backend",
			users:    []domain.User{},
			teams: []domain.Team{
				domain.NewTeam("backend", []domain.User{}),
			},
			prs: []domain.PullRequest{},
			verify: func(t *testing.T, result DeactivateResult, userStorage *fakeUserStorage, prStorage *fakePullRequestStorage) {
				t.Helper()
				if result.DeactivatedCount != 0 {
					t.Fatalf("expected 0 deactivations, got %d", result.DeactivatedCount)
				}
				if result.ReassignedPRCount != 0 {
					t.Fatalf("expected 0 reassignments, got %d", result.ReassignedPRCount)
				}
			},
		},
		{
			name:     "all users already inactive",
			teamName: "backend",
			users: []domain.User{
				domain.NewUser("u1", "Alice", "backend", false),
				domain.NewUser("u2", "Bob", "backend", false),
			},
			teams: []domain.Team{
				domain.NewTeam("backend", []domain.User{
					domain.NewUser("u1", "Alice", "backend", false),
					domain.NewUser("u2", "Bob", "backend", false),
				}),
			},
			prs: []domain.PullRequest{},
			verify: func(t *testing.T, result DeactivateResult, userStorage *fakeUserStorage, prStorage *fakePullRequestStorage) {
				t.Helper()
				if result.DeactivatedCount != 0 {
					t.Fatalf("expected 0 deactivations, got %d", result.DeactivatedCount)
				}
			},
		},
		{
			name:     "reassigns PRs when replacement available",
			teamName: "backend",
			users: []domain.User{
				domain.NewUser("u1", "Alice", "backend", true),
				domain.NewUser("u2", "Bob", "backend", true),
				domain.NewUser("u3", "Charlie", "backend", true),
				domain.NewUser("u4", "Dave", "backend", true),
			},
			teams: []domain.Team{
				domain.NewTeam("backend", []domain.User{
					domain.NewUser("u1", "Alice", "backend", true),
					domain.NewUser("u2", "Bob", "backend", true),
					domain.NewUser("u3", "Charlie", "backend", true),
					domain.NewUser("u4", "Dave", "backend", true),
				}),
			},
			prs: []domain.PullRequest{
				func() domain.PullRequest {
					pr := domain.NewPullRequest("pr-1", "Fix", "u1", "backend", time.Now())
					pr.AssignReviewers([]string{"u2"})
					return pr
				}(),
			},
			verify: func(t *testing.T, result DeactivateResult, userStorage *fakeUserStorage, prStorage *fakePullRequestStorage) {
				t.Helper()
				if result.ReassignedPRCount != 1 {
					t.Fatalf("expected 1 reassignment, got %d", result.ReassignedPRCount)
				}
				pr, _ := prStorage.GetPullRequest(context.Background(), "pr-1")
				if len(pr.Reviewers) == 0 || pr.Reviewers[0] == "u2" {
					t.Fatalf("expected u2 to be replaced, got reviewers %v", pr.Reviewers)
				}
			},
		},
		{
			name:     "removes reviewer when no replacement available",
			teamName: "backend",
			users: []domain.User{
				domain.NewUser("u1", "Alice", "backend", true),
				domain.NewUser("u2", "Bob", "backend", true),
			},
			teams: []domain.Team{
				domain.NewTeam("backend", []domain.User{
					domain.NewUser("u1", "Alice", "backend", true),
					domain.NewUser("u2", "Bob", "backend", true),
				}),
			},
			prs: []domain.PullRequest{
				func() domain.PullRequest {
					pr := domain.NewPullRequest("pr-1", "Fix", "u1", "backend", time.Now())
					pr.AssignReviewers([]string{"u2"})
					return pr
				}(),
			},
			verify: func(t *testing.T, result DeactivateResult, userStorage *fakeUserStorage, prStorage *fakePullRequestStorage) {
				t.Helper()
				pr, _ := prStorage.GetPullRequest(context.Background(), "pr-1")
				if len(pr.Reviewers) != 0 {
					t.Fatalf("expected reviewer removed, got %v", pr.Reviewers)
				}
			},
		},
		{
			name:     "skips merged PRs",
			teamName: "backend",
			users: []domain.User{
				domain.NewUser("u1", "Alice", "backend", true),
				domain.NewUser("u2", "Bob", "backend", true),
			},
			teams: []domain.Team{
				domain.NewTeam("backend", []domain.User{
					domain.NewUser("u1", "Alice", "backend", true),
					domain.NewUser("u2", "Bob", "backend", true),
				}),
			},
			prs: []domain.PullRequest{
				func() domain.PullRequest {
					pr := domain.NewPullRequest("pr-1", "Fix", "u1", "backend", time.Now())
					pr.AssignReviewers([]string{"u2"})
					pr.Status = "MERGED"
					return pr
				}(),
			},
			verify: func(t *testing.T, result DeactivateResult, userStorage *fakeUserStorage, prStorage *fakePullRequestStorage) {
				t.Helper()
				if result.ReassignedPRCount != 0 {
					t.Fatalf("expected 0 reassignments for merged PR, got %d", result.ReassignedPRCount)
				}
			},
		},
		{
			name:     "team not found",
			teamName: "nonexistent",
			wantErr:  domain.ErrTeamNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userStorage := newFakeUserStorage(tt.users...)
			teamStorage := newFakeTeamStorage(tt.teams...)
			prStorage := newFakePullRequestStorage(tt.prs...)

			uc := NewDeactivateTeamUsersUseCase(userStorage, teamStorage, prStorage, &fakeRandom{}, testLogger())

			result, err := uc.DeactivateTeamUsers(ctx, tt.teamName)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if tt.wantErr == nil && tt.verify != nil {
				tt.verify(t, result, userStorage, prStorage)
			}
		})
	}
}

// test helpers

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

type fakeUserStorage struct {
	users       map[string]domain.User
	createErr   error
	createErrID string
	getErr      error
	getErrID    string
	updateErr   error
	updateErrID string
	listErr     error
}

func newFakeUserStorage(users ...domain.User) *fakeUserStorage {
	m := make(map[string]domain.User, len(users))
	for _, u := range users {
		m[u.ID] = u
	}
	return &fakeUserStorage{users: m}
}

func (f *fakeUserStorage) CreateUser(_ context.Context, user domain.User) error {
	if f.createErr != nil && (f.createErrID == "" || f.createErrID == user.ID) {
		return f.createErr
	}
	f.users[user.ID] = user
	return nil
}

func (f *fakeUserStorage) ListUsers(_ context.Context) ([]domain.User, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	result := make([]domain.User, 0, len(f.users))
	for _, user := range f.users {
		result = append(result, user)
	}
	return result, nil
}

func (f *fakeUserStorage) GetUser(_ context.Context, id string) (domain.User, error) {
	if f.getErr != nil && (f.getErrID == "" || f.getErrID == id) {
		return domain.User{}, f.getErr
	}
	user, ok := f.users[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

func (f *fakeUserStorage) UpdateUser(_ context.Context, user domain.User) error {
	if _, ok := f.users[user.ID]; !ok {
		return domain.ErrUserNotFound
	}
	if f.updateErr != nil && (f.updateErrID == "" || f.updateErrID == user.ID) {
		return f.updateErr
	}
	f.users[user.ID] = user
	return nil
}

type fakeTeamStorage struct {
	teams         map[string]domain.Team
	getErr        error
	getErrName    string
	createErr     error
	createErrName string
}

func newFakeTeamStorage(teams ...domain.Team) *fakeTeamStorage {
	m := make(map[string]domain.Team, len(teams))
	for _, team := range teams {
		m[team.Name] = team
	}
	return &fakeTeamStorage{teams: m}
}

func (f *fakeTeamStorage) CreateTeam(_ context.Context, team domain.Team) error {
	if f.createErr != nil && (f.createErrName == "" || f.createErrName == team.Name) {
		return f.createErr
	}
	if _, exists := f.teams[team.Name]; exists {
		return domain.ErrTeamExists
	}
	f.teams[team.Name] = team
	return nil
}

func (f *fakeTeamStorage) ListTeams(_ context.Context) ([]domain.Team, error) {
	result := make([]domain.Team, 0, len(f.teams))
	for _, team := range f.teams {
		result = append(result, team)
	}
	return result, nil
}

func (f *fakeTeamStorage) GetTeam(_ context.Context, name string) (domain.Team, error) {
	if f.getErr != nil && (f.getErrName == "" || f.getErrName == name) {
		return domain.Team{}, f.getErr
	}
	team, ok := f.teams[name]
	if !ok {
		return domain.Team{}, domain.ErrTeamNotFound
	}
	return team, nil
}

type fakePullRequestStorage struct {
	prs               map[string]domain.PullRequest
	listErr           error
	listByReviewerErr error
	createErr         error
	createErrID       string
	getErr            error
	getErrID          string
	updateErr         error
	updateErrID       string
}

func newFakePullRequestStorage(prs ...domain.PullRequest) *fakePullRequestStorage {
	m := make(map[string]domain.PullRequest, len(prs))
	for _, pr := range prs {
		m[pr.ID] = pr
	}
	return &fakePullRequestStorage{prs: m}
}

func (f *fakePullRequestStorage) CreatePullRequest(_ context.Context, pr domain.PullRequest) error {
	if f.createErr != nil && (f.createErrID == "" || f.createErrID == pr.ID) {
		return f.createErr
	}
	if _, exists := f.prs[pr.ID]; exists {
		return domain.ErrPullRequestExists
	}
	f.prs[pr.ID] = pr
	return nil
}

func (f *fakePullRequestStorage) ListPullRequests(_ context.Context) ([]domain.PullRequest, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	result := make([]domain.PullRequest, 0, len(f.prs))
	for _, pr := range f.prs {
		result = append(result, pr)
	}
	return result, nil
}

func (f *fakePullRequestStorage) GetPullRequest(_ context.Context, id string) (domain.PullRequest, error) {
	if f.getErr != nil && (f.getErrID == "" || f.getErrID == id) {
		return domain.PullRequest{}, f.getErr
	}
	pr, ok := f.prs[id]
	if !ok {
		return domain.PullRequest{}, domain.ErrPullRequestNotFound
	}
	return pr, nil
}

func (f *fakePullRequestStorage) UpdatePullRequest(_ context.Context, pr domain.PullRequest) error {
	if _, ok := f.prs[pr.ID]; !ok {
		return domain.ErrPullRequestNotFound
	}
	if f.updateErr != nil && (f.updateErrID == "" || f.updateErrID == pr.ID) {
		return f.updateErr
	}
	f.prs[pr.ID] = pr
	return nil
}

func (f *fakePullRequestStorage) ListPullRequestsByReviewer(_ context.Context, reviewerID string) ([]domain.PullRequest, error) {
	if f.listByReviewerErr != nil {
		return nil, f.listByReviewerErr
	}
	result := make([]domain.PullRequest, 0)
	for _, pr := range f.prs {
		for _, reviewer := range pr.Reviewers {
			if reviewer == reviewerID {
				result = append(result, pr)
				break
			}
		}
	}
	return result, nil
}

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time {
	return f.now
}

type fakeRandom struct {
	calls int
}

func (f *fakeRandom) Shuffle(n int, swap func(i, j int)) {
	f.calls++
	if n > 1 {
		swap(0, n-1)
	}
}

func stringPtr(s string) *string {
	return &s
}
