package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ayushverma21-dev/Student-api/internal/config"
	"github.com/ayushverma21-dev/Student-api/internal/storage"
	"github.com/ayushverma21-dev/Student-api/internal/types"
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

func (storage *Postgres) GetStudentById(id int64) (types.Student, error) {
	var student types.Student
	err := storage.DB.QueryRow(
		context.Background(),
		`SELECT id, name, email, age FROM public.students WHERE id = $1`,
		id,
	).Scan(&student.Id, &student.Name, &student.Email, &student.Age)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return types.Student{}, fmt.Errorf("no student found with id %d", id)
		}
		return types.Student{}, fmt.Errorf("query student: %w", err)
	}

	return student, nil
}

func (db *Postgres) UpdateStudent(id int64, name string, email string, age int) error {
	result, err := db.DB.Exec(
		context.Background(),
		`UPDATE public.students SET name = $1, email = $2, age = $3 WHERE id = $4`,
		name,
		email,
		age,
		id,
	)
	if err != nil {
		return fmt.Errorf("update student: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("student %d: %w", id, storage.ErrStudentNotFound)
	}

	return nil
}

func (db *Postgres) DeleteStudent(id int64) error {
	result, err := db.DB.Exec(
		context.Background(),
		`DELETE FROM public.students WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete student: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("student %d: %w", id, storage.ErrStudentNotFound)
	}

	return nil
}

func (storage *Postgres) Close() {
	storage.DB.Close(context.Background())
}

func (storage *Postgres) GetStudents() ([]types.Student, error) {
	rows, err := storage.DB.Query(
		context.Background(),
		`SELECT id, name, email, age FROM public.students ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("query students: %w", err)
	}
	defer rows.Close()

	var students []types.Student
	for rows.Next() {
		var student types.Student
		if err := rows.Scan(&student.Id, &student.Name, &student.Email, &student.Age); err != nil {
			return nil, fmt.Errorf("scan student: %w", err)
		}
		students = append(students, student)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate students: %w", err)
	}

	return students, nil
}
