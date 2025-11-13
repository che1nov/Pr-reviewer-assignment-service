package postgresql

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const (
	connTimeout        = 5 * time.Second
	maxOpenConnections = 10
	maxIdleConnections = 5
	connLifeTime       = 5 * time.Minute
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func NewConnection(dsn string, log *slog.Logger) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		log.Error("не удалось открыть подключение PostgreSQL", "error", err)
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConnections)
	db.SetMaxIdleConns(maxIdleConnections)
	db.SetConnMaxLifetime(connLifeTime)

	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Error("не удалось выполнить ping PostgreSQL", "error", err)
		_ = db.Close()
		return nil, err
	}

	log.Info("подключение к PostgreSQL успешно установлено")
	return db, nil
}

func RunMigrations(db *sql.DB, log *slog.Logger) error {
	source, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		log.Error("не удалось подготовить источник миграций", "error", err)
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Error("не удалось создать драйвер миграций", "error", err)
		return err
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		log.Error("не удалось инициализировать миграции", "error", err)
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error("ошибка применения миграций", "error", err)
		return err
	}

	log.Info("миграции выполнены успешно")
	return nil
}
