//go:generate minimock -i .UserService,.AuthService -o mocks -s _mock.go -g
package loginhandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/hydra13/gophermart/internal/models"
	"github.com/hydra13/gophermart/internal/validators"
)

type JSONRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserService interface {
	Login(ctx context.Context, login, password string) (int64, error)
}

type AuthService interface {
	SetAuthCookie(w http.ResponseWriter, userID int64)
}

type Handler struct {
	log zerolog.Logger
	u   UserService
	a   AuthService
}

func NewHandler(u UserService, a AuthService, log zerolog.Logger) *Handler {
	return &Handler{
		u:   u,
		a:   a,
		log: log,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		h.log.Debug().
			Str("content-type", r.Header.Get("Content-Type")).
			Msg("login: error by content-type")

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req JSONRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.log.Debug().
			Err(err).
			Msg("login: error read request body")

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !validators.IsValidEmail(req.Login) || !validators.IsValidPass(req.Password) {
		h.log.Debug().
			Str("login", req.Login).
			Str("password", req.Password).
			Msg("login: error by login or password")

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.u.Login(r.Context(), req.Login, req.Password)

	if err != nil {
		if err == models.ErrUserNotFound {
			h.log.Debug().
				Err(err).
				Msg("login: invalid login or password")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.log.Error().
			Err(err).
			Msg("login: error login user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.a.SetAuthCookie(w, userID)

	w.WriteHeader(http.StatusOK)
}
