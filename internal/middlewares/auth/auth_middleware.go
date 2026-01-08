package auth

import (
	"net/http"

	"github.com/rs/zerolog"

	authContext "github.com/hydra13/gophermart/internal/services/auth_context"
)

type AuthService interface {
	GetUser(request *http.Request) (userID int64, err error)
	SetAuthCookie(w http.ResponseWriter, userID int64)
}

func NewAuthMiddleware(as AuthService, log zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := as.GetUser(r)
			if err != nil {
				log.Debug().
					Err(err).
					Msg("auth middleware: can't get user from cookie")

				w.WriteHeader(http.StatusUnauthorized)

				return
			}

			log.Debug().
				Int64("user_id", userID).
				Msg("auth middleware")

			ctx := authContext.CreateContextWithUserID(r.Context(), userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
