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
	"golang.org/x/crypto/bcrypt"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Postgres struct {
	DB *gorm.DB
}

type studentModel struct {
	Id           int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name         string `gorm:"column:name;not null"`
	Email        string `gorm:"column:email;not null"`
	PasswordHash string `gorm:"column:password_hash;not null;default:''"`
	Age          int    `gorm:"column:age;not null;check:age > 0"`
}

func (studentModel) TableName() string {
	return "public.students"
}

func New(cfg *config.Config) (*Postgres, error) {
	ctx := context.Background()
	connection, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "3D000" {
			return nil, err
		}

		if err := createDatabase(ctx, cfg.DatabaseURL); err != nil {
			return nil, err
		}
		connection, err = pgx.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
	}
	connection.Close(ctx)

	db, err := gorm.Open(postgresDriver.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&studentModel{}); err != nil {
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

func (storage *Postgres) CreateStudent(name string, email string, password string, age int) (int64, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}

	student := studentModel{
		Name:         name,
		Email:        email,
		PasswordHash: string(passwordHash),
		Age:          age,
	}
	err = storage.DB.Create(&student).Error
	if err != nil {
		return 0, fmt.Errorf("create student: %w", err)
	}

	return student.Id, nil
}

func (db *Postgres) AuthenticateStudent(email string, password string) (types.Student, error) {
	var model studentModel
	err := db.DB.Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.Student{}, storage.ErrInvalidCredentials
		}
		return types.Student{}, fmt.Errorf("find student for login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(model.PasswordHash), []byte(password)); err != nil {
		return types.Student{}, storage.ErrInvalidCredentials
	}
	return toStudent(model), nil
}

func (storage *Postgres) GetStudentById(id int64) (types.Student, error) {
	var model studentModel
	err := storage.DB.First(&model, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.Student{}, fmt.Errorf("no student found with id %d", id)
		}
		return types.Student{}, fmt.Errorf("query student: %w", err)
	}

	return toStudent(model), nil
}

func (db *Postgres) UpdateStudent(id int64, name string, email string, age int) error {
	result := db.DB.Model(&studentModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":  name,
		"email": email,
		"age":   age,
	})
	if result.Error != nil {
		return fmt.Errorf("update student: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("student %d: %w", id, storage.ErrStudentNotFound)
	}

	return nil
}

func (db *Postgres) DeleteStudent(id int64) error {
	result := db.DB.Delete(&studentModel{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete student: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("student %d: %w", id, storage.ErrStudentNotFound)
	}

	return nil
}

func (storage *Postgres) Close() {
	sqlDB, err := storage.DB.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}

func (storage *Postgres) GetStudents() ([]types.Student, error) {
	var models []studentModel
	if err := storage.DB.Order("id").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("query students: %w", err)
	}

	students := make([]types.Student, 0, len(models))
	for _, model := range models {
		students = append(students, toStudent(model))
	}
	return students, nil
}

func toStudent(model studentModel) types.Student {
	return types.Student{
		Id:    model.Id,
		Name:  model.Name,
		Email: model.Email,
		Age:   model.Age,
	}
}
