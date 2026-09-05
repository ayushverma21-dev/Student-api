package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ayushverma21-dev/Student-api/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Postgres struct {
	DB *pgx.Conn
}

func New(cfg *config.Config) (*Postgres, error) {
	ctx := context.Background()
	db, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "3D000" {
			return nil, err
		}

		if err := createDatabase(ctx, cfg.DatabaseURL); err != nil {
			return nil, err
		}
		db, err = pgx.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
	}

	if err := db.Ping(ctx); err != nil {
		db.Close(ctx)
		return nil, err
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS public.students (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			age INTEGER NOT NULL CHECK (age > 0)
		)
	`)
	if err != nil {
		db.Close(ctx)
		return nil, err
	}
	slog.Info("database table initialized", slog.String("table", "public.students"))

	return &Postgres{DB: db}, nil
}

func createDatabase(ctx context.Context, databaseURL string) error {
	connectionConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return err
	}

	databaseName := connectionConfig.Database
	connectionConfig.Database = "postgres"
	admin, err := pgx.ConnectConfig(ctx, connectionConfig)
	if err != nil {
		return err
	}
	defer admin.Close(ctx)

	_, err = admin.Exec(ctx, "CREATE DATABASE "+quoteIdentifier(databaseName))
	return err
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func (storage *Postgres) CreateStudent(name string, email string, age int) (int64, error) {
	var id int64
	err := storage.DB.QueryRow(
		context.Background(),
		`INSERT INTO public.students (name, email, age) VALUES ($1, $2, $3) RETURNING id`,
		name,
		email,
		age,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create student: %w", err)
	}

	return id, nil
}

func (storage *Postgres) Close() {
	storage.DB.Close(context.Background())
}
