package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ayushverma21-dev/Student-api/internal/redis"
	"github.com/ayushverma21-dev/Student-api/internal/storage"
	"github.com/ayushverma21-dev/Student-api/internal/types"
	"github.com/ayushverma21-dev/Student-api/internal/utils/response"
	"github.com/go-playground/validator/v10"
)

func Login(store storage.Storage, sessions *redis.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var credentials types.StudentCredentials
		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}
		if err := validator.New().Struct(credentials); err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(err.(validator.ValidationErrors)))
			return
		}

		student, err := store.AuthenticateStudent(credentials.Email, credentials.Password)
		if err != nil {
			if errors.Is(err, storage.ErrInvalidCredentials) {
				response.WriteJson(w, http.StatusUnauthorized, response.GeneralError(err))
				return
			}
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		token, err := sessions.CreateSession(r.Context(), student)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}
		response.WriteJson(w, http.StatusOK, map[string]string{"token": token})
	}
}
