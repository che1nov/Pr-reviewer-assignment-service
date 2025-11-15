package postgresql

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/jmoiron/sqlx"

	"github.com/che1nov/Pr-reviewer-assignment-service/internal/domain"
)

type TeamAdapter struct {
	db  *sqlx.DB
	log *slog.Logger
}

func NewTeamAdapter(db *sqlx.DB, log *slog.Logger) *TeamAdapter {
	return &TeamAdapter{
		db:  db,
		log: log,
	}
}

// CreateTeam сохраняет команду.
func (a *TeamAdapter) CreateTeam(ctx context.Context, team domain.Team) error {
	const query = `
		INSERT INTO teams (name)
		VALUES ($1)
	`

	if _, err := a.db.ExecContext(ctx, query, team.Name); err != nil {
		a.log.ErrorContext(ctx, "ошибка создания команды", "team_name", team.Name, "error", err)
		return err
	}

	return nil
}

// ListTeams возвращает список команд.
func (a *TeamAdapter) ListTeams(ctx context.Context) ([]domain.Team, error) {
	const query = `
		SELECT name
		FROM teams
		ORDER BY name
	`

	var names []string
	if err := a.db.SelectContext(ctx, &names, query); err != nil {
		a.log.ErrorContext(ctx, "ошибка получения списка команд", "error", err)
		return nil, err
	}

	teams := make([]domain.Team, 0, len(names))
	for _, name := range names {
		team, err := a.GetTeam(ctx, name)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, nil
}

// GetTeam возвращает команду по имени.
func (a *TeamAdapter) GetTeam(ctx context.Context, name string) (domain.Team, error) {
	const queryTeam = `
		SELECT name
		FROM teams
		WHERE name = $1
	`

	var teamName string
	if err := a.db.GetContext(ctx, &teamName, queryTeam, name); err != nil {
		if err == sql.ErrNoRows {
			return domain.Team{}, domain.ErrTeamNotFound
		}
		a.log.ErrorContext(ctx, "ошибка получения команды", "team_name", name, "error", err)
		return domain.Team{}, err
	}

	const queryUsers = `
		SELECT id, name, team_name, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY id
	`

	var members []domain.User
	if err := a.db.SelectContext(ctx, &members, queryUsers, teamName); err != nil {
		a.log.ErrorContext(ctx, "ошибка получения участников команды", "team_name", name, "error", err)
		return domain.Team{}, err
	}

	return domain.NewTeam(teamName, members), nil
}
