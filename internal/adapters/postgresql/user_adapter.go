package postgresql

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/jmoiron/sqlx"

	"github.com/che1nov/Pr-reviewer-assignment-service/internal/domain"
)

type UserAdapter struct {
	db  *sqlx.DB
	log *slog.Logger
}

func NewUserAdapter(db *sqlx.DB, log *slog.Logger) *UserAdapter {
	return &UserAdapter{
		db:  db,
		log: log,
	}
}

// CreateUser сохраняет нового пользователя.
func (a *UserAdapter) CreateUser(ctx context.Context, user domain.User) error {
	const query = `
		INSERT INTO users (id, name, team_name, is_active)
		VALUES ($1, $2, $3, $4)
	`

	if _, err := a.db.ExecContext(ctx, query, user.ID, user.Name, user.TeamName, user.IsActive); err != nil {
		a.log.ErrorContext(ctx, "ошибка создания пользователя", "user_id", user.ID, "error", err)
		return err
	}

	return nil
}

// ListUsers возвращает список пользователей.
func (a *UserAdapter) ListUsers(ctx context.Context) ([]domain.User, error) {
	const query = `
		SELECT id, name, team_name, is_active
		FROM users
		ORDER BY id
	`

	var users []domain.User
	if err := a.db.SelectContext(ctx, &users, query); err != nil {
		a.log.ErrorContext(ctx, "ошибка получения списка пользователей", "error", err)
		return nil, err
	}

	return users, nil
}

// GetUser возвращает пользователя по идентификатору.
func (a *UserAdapter) GetUser(ctx context.Context, id string) (domain.User, error) {
	const query = `
		SELECT id, name, team_name, is_active
		FROM users
		WHERE id = $1
	`

	var user domain.User
	if err := a.db.GetContext(ctx, &user, query, id); err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, domain.ErrUserNotFound
		}
		a.log.ErrorContext(ctx, "ошибка получения пользователя", "user_id", id, "error", err)
		return domain.User{}, err
	}

	return user, nil
}

// UpdateUser обновляет пользователя.
func (a *UserAdapter) UpdateUser(ctx context.Context, user domain.User) error {
	const query = `
		UPDATE users
		SET name = $2, team_name = $3, is_active = $4
		WHERE id = $1
	`

	result, err := a.db.ExecContext(ctx, query, user.ID, user.Name, user.TeamName, user.IsActive)
	if err != nil {
		a.log.ErrorContext(ctx, "ошибка обновления пользователя", "user_id", user.ID, "error", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		a.log.ErrorContext(ctx, "ошибка чтения результата обновления пользователя", "user_id", user.ID, "error", err)
		return err
	}
	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
