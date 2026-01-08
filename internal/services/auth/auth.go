package auth

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hydra13/gophermart/internal/models"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

const (
	cookieTokenExp = time.Hour * 3
	cookieKey      = "t"
	secretKey      = "supersecretkey"
)

type AuthService struct{}

func New() *AuthService {
	return &AuthService{}
}

func (a *AuthService) GetUser(r *http.Request) (userID int64, err error) {
	cookie, err := r.Cookie(cookieKey)
	if err != nil {
		return -1, models.ErrTokenNotFound
	}

	return a.getUserIDFromToken(cookie.Value)
}

func (a *AuthService) SetAuthCookie(w http.ResponseWriter, userID int64) {
	tokenString, err := a.BuildJWTString(userID)
	if err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:     cookieKey,
			Value:    tokenString,
			Expires:  time.Now().Add(cookieTokenExp),
			HttpOnly: true,
			// Secure:   true,
			// SameSite: http.SameSiteStrictMode,
			SameSite: http.SameSiteNoneMode,
		})
	}
}

func (a *AuthService) BuildJWTString(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cookieTokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *AuthService) getUserIDFromToken(tokenString string) (int64, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return -1, err
	}

	if !token.Valid {
		return -1, models.ErrTokenNotValid
	}

	return claims.UserID, nil
}
