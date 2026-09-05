package storage

import (
	"errors"

	"github.com/ayushverma21-dev/Student-api/internal/types"
)

var ErrStudentNotFound = errors.New("student not found")

type Storage interface {
	CreateStudent(name string, email string, age int) (int64, error)
	UpdateStudent(id int64, name string, email string, age int) error
	DeleteStudent(id int64) error
	GetStudentById(id int64) (types.Student, error)
	GetStudents() ([]types.Student, error)
}
