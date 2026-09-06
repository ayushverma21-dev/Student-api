package storage

import (
	"errors"

	"github.com/ayushverma21-dev/Student-api/internal/types"
)

var ErrStudentNotFound = errors.New("student not found")
var ErrInvalidCredentials = errors.New("invalid credentials")

type Storage interface {
	CreateStudent(name string, email string, password string, age int) (int64, error)
	AuthenticateStudent(email string, password string) (types.Student, error)
	UpdateStudent(id int64, name string, email string, age int) error
	DeleteStudent(id int64) error
	GetStudentById(id int64) (types.Student, error)
	GetStudents() ([]types.Student, error)
}
