package postgresql

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/jmoiron/sqlx"

	"github.com/che1nov/Pr-reviewer-assignment-service/internal/domain"
)

type PullRequestAdapter struct {
	db  *sqlx.DB
	log *slog.Logger
}

func NewPullRequestAdapter(db *sqlx.DB, log *slog.Logger) *PullRequestAdapter {
	return &PullRequestAdapter{
		db:  db,
		log: log,
	}
}

// CreatePullRequest сохраняет новый pull request.
func (a *PullRequestAdapter) CreatePullRequest(ctx context.Context, pr domain.PullRequest) error {
	const query = `
		INSERT INTO pull_requests (id, title, author_id, team_name, status, created_at, merged_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	if _, err := a.db.ExecContext(ctx, query, pr.ID, pr.Title, pr.AuthorID, pr.TeamName, pr.Status, pr.CreatedAt, pr.MergedAt); err != nil {
		a.log.ErrorContext(ctx, "ошибка создания pull request", "pr_id", pr.ID, "error", err)
		return err
	}

	if err := a.replaceReviewers(ctx, pr.ID, pr.Reviewers); err != nil {
		return err
	}

	return nil
}

// ListPullRequests возвращает все pull request.
func (a *PullRequestAdapter) ListPullRequests(ctx context.Context) ([]domain.PullRequest, error) {
	const query = `
		SELECT id, title, author_id, team_name, status, created_at, merged_at
		FROM pull_requests
		ORDER BY created_at DESC
	`

	rows, err := a.db.QueryxContext(ctx, query)
	if err != nil {
		a.log.ErrorContext(ctx, "ошибка получения списка pull request", "error", err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []domain.PullRequest
	for rows.Next() {
		pr, err := scanPullRequest(rows)
		if err != nil {
			a.log.ErrorContext(ctx, "ошибка чтения pull request", "error", err)
			return nil, err
		}

		reviewers, err := a.loadReviewers(ctx, pr.ID)
		if err != nil {
			return nil, err
		}
		pr.Reviewers = reviewers

		result = append(result, pr)
	}

	return result, nil
}

// GetPullRequest получает pull request по идентификатору.
func (a *PullRequestAdapter) GetPullRequest(ctx context.Context, id string) (domain.PullRequest, error) {
	const query = `
		SELECT id, title, author_id, team_name, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1
	`

	row := a.db.QueryRowxContext(ctx, query, id)
	pr, err := scanPullRequest(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.PullRequest{}, domain.ErrPullRequestNotFound
		}
		a.log.ErrorContext(ctx, "ошибка получения pull request", "pr_id", id, "error", err)
		return domain.PullRequest{}, err
	}

	reviewers, err := a.loadReviewers(ctx, id)
	if err != nil {
		return domain.PullRequest{}, err
	}
	pr.Reviewers = reviewers

	return pr, nil
}

// UpdatePullRequest обновляет данные pull request.
func (a *PullRequestAdapter) UpdatePullRequest(ctx context.Context, pr domain.PullRequest) error {
	const query = `
		UPDATE pull_requests
		SET title = $2,
			author_id = $3,
			team_name = $4,
			status = $5,
			created_at = $6,
			merged_at = $7
		WHERE id = $1
	`

	result, err := a.db.ExecContext(ctx, query, pr.ID, pr.Title, pr.AuthorID, pr.TeamName, pr.Status, pr.CreatedAt, pr.MergedAt)
	if err != nil {
		a.log.ErrorContext(ctx, "ошибка обновления pull request", "pr_id", pr.ID, "error", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		a.log.ErrorContext(ctx, "ошибка чтения результата обновления pull request", "pr_id", pr.ID, "error", err)
		return err
	}
	if rows == 0 {
		return domain.ErrPullRequestNotFound
	}

	if err := a.replaceReviewers(ctx, pr.ID, pr.Reviewers); err != nil {
		return err
	}

	return nil
}

// ListPullRequestsByReviewer возвращает pull request пользователя.
func (a *PullRequestAdapter) ListPullRequestsByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	const query = `
		SELECT pr.id, pr.title, pr.author_id, pr.team_name, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		JOIN pull_request_reviewers r ON pr.id = r.pr_id
		WHERE r.reviewer_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := a.db.QueryxContext(ctx, query, reviewerID)
	if err != nil {
		a.log.ErrorContext(ctx, "ошибка получения pull request по ревьюеру", "reviewer_id", reviewerID, "error", err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []domain.PullRequest
	for rows.Next() {
		pr, err := scanPullRequest(rows)
		if err != nil {
			a.log.ErrorContext(ctx, "ошибка чтения pull request для ревьюера", "reviewer_id", reviewerID, "error", err)
			return nil, err
		}

		reviewers, err := a.loadReviewers(ctx, pr.ID)
		if err != nil {
			return nil, err
		}
		pr.Reviewers = reviewers

		result = append(result, pr)
	}

	return result, nil
}

func (a *PullRequestAdapter) loadReviewers(ctx context.Context, prID string) ([]string, error) {
	const query = `
		SELECT reviewer_id
		FROM pull_request_reviewers
		WHERE pr_id = $1
		ORDER BY reviewer_id
	`

	rows, err := a.db.QueryxContext(ctx, query, prID)
	if err != nil {
		a.log.ErrorContext(ctx, "ошибка получения ревьюеров pull request", "pr_id", prID, "error", err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	reviewers := make([]string, 0, 2)
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			a.log.ErrorContext(ctx, "ошибка чтения ревьюера", "pr_id", prID, "error", err)
			return nil, err
		}
		reviewers = append(reviewers, reviewerID)
	}

	return reviewers, nil
}

func (a *PullRequestAdapter) replaceReviewers(ctx context.Context, prID string, reviewers []string) error {
	const deleteQuery = `
		DELETE FROM pull_request_reviewers
		WHERE pr_id = $1
	`

	if _, err := a.db.ExecContext(ctx, deleteQuery, prID); err != nil {
		a.log.ErrorContext(ctx, "ошибка очистки ревьюеров pull request", "pr_id", prID, "error", err)
		return err
	}

	if len(reviewers) == 0 {
		return nil
	}

	const insertQuery = `
		INSERT INTO pull_request_reviewers (pr_id, reviewer_id)
		VALUES ($1, $2)
	`

	for _, reviewerID := range reviewers {
		if _, err := a.db.ExecContext(ctx, insertQuery, prID, reviewerID); err != nil {
			a.log.ErrorContext(ctx, "ошибка сохранения ревьюера pull request", "pr_id", prID, "reviewer_id", reviewerID, "error", err)
			return err
		}
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPullRequest(scanner rowScanner) (domain.PullRequest, error) {
	var (
		pr       domain.PullRequest
		mergedAt sql.NullTime
	)

	if err := scanner.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.TeamName, &pr.Status, &pr.CreatedAt, &mergedAt); err != nil {
		return domain.PullRequest{}, err
	}
	if mergedAt.Valid {
		t := mergedAt.Time
		pr.MergedAt = &t
	}

	return pr, nil
}
