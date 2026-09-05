package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ayushverma21-dev/Student-api/internal/storage"
	"github.com/ayushverma21-dev/Student-api/internal/types"
	"github.com/ayushverma21-dev/Student-api/internal/utils/response"
	"github.com/go-playground/validator/v10"
)

func New(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("creating a student")
		var student types.Student
		err := json.NewDecoder(r.Body).Decode((&student))
		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("Empty body")))

			return
		}
		// _, _ = w.Write([]byte("Welcomme to students api"))

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		// request validation

		if err := validator.New().Struct(student); err != nil {
			validateErrs := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}
		lastId, err := storage.CreateStudent(
			student.Name,
			student.Email,
			student.Age,
		)
		slog.Info("User Created Successfully", slog.String("userrId", fmt.Sprint(lastId)))

		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, err)
			return
		}
		response.WriteJson(w, http.StatusCreated, map[string]int64{"id": lastId})
	}
}

func GetById(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("geting a student", slog.String("id", id))
		intId, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}
		student, err := storage.GetStudentById(intId)

		if err != nil {
			slog.Error("error getting user", slog.String("id", id))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}
		response.WriteJson(w, http.StatusOK, student)

	}
}

func Update(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		var student types.Student
		if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if err := validator.New().Struct(student); err != nil {
			validateErrs := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}

		if err := store.UpdateStudent(id, student.Name, student.Email, student.Age); err != nil {
			if errors.Is(err, storage.ErrStudentNotFound) {
				slog.Warn("student not found for update", slog.Int64("id", id))
				response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
				return
			}
			slog.Error("failed to update student", slog.Int64("id", id), slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		student.Id = id
		slog.Info("student updated successfully", slog.Int64("id", id))
		response.WriteJson(w, http.StatusOK, student)
	}
}

func Delete(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if err := store.DeleteStudent(id); err != nil {
			if errors.Is(err, storage.ErrStudentNotFound) {
				slog.Warn("student not found for deletion", slog.Int64("id", id))
				response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
				return
			}
			slog.Error("failed to delete student", slog.Int64("id", id), slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		slog.Info("student deleted successfully", slog.Int64("id", id))
		w.WriteHeader(http.StatusNoContent)
	}
}

func GetList(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("getting all students")

		students, err := storage.GetStudents()
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, err)
			return
		}
		response.WriteJson(w, http.StatusOK, students)

	}
}
