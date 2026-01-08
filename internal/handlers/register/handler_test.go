package registerhandler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/hydra13/gophermart/internal/handlers/register/mocks"
	"github.com/hydra13/gophermart/internal/models"
)

type request struct {
	contentType string
	body        string
}

func TestHandler_Handle(t *testing.T) {
	tests := []struct {
		name        string
		request     request
		userService func(mc *minimock.Controller) UserService
		authService func(mc *minimock.Controller) AuthService
		expected    int
	}{
		{
			name: "success",
			request: request{
				contentType: "application/json",
				body:        `{"login":"testuser@email.com","password":"testPass123"}`,
			},
			userService: func(mc *minimock.Controller) UserService {
				return mocks.NewUserServiceMock(mc).
					RegisterMock.
					Expect(context.Background(), "testuser@email.com", "testPass123").
					Return(int64(1234567890), nil)
			},
			authService: func(mc *minimock.Controller) AuthService {
				return mocks.NewAuthServiceMock(mc).
					SetAuthCookieMock.
					ExpectUserIDParam2(int64(1234567890)).
					Return()
			},
			expected: http.StatusOK,
		},
		{
			name: "invalid content type",
			request: request{
				contentType: "text/plain",
				body:        `{"login":"testuser@email.com","password":"testPass123"}`,
			},
			userService: func(mc *minimock.Controller) UserService {
				return mocks.NewUserServiceMock(mc)
			},
			authService: func(mc *minimock.Controller) AuthService {
				return mocks.NewAuthServiceMock(mc)
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "invalid json",
			request: request{
				contentType: "application/json",
				body:        `{"login":"testuser@email.com","password":"testPass123"`,
			},
			userService: func(mc *minimock.Controller) UserService {
				return mocks.NewUserServiceMock(mc)
			},
			authService: func(mc *minimock.Controller) AuthService {
				return mocks.NewAuthServiceMock(mc)
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "invalid login",
			request: request{
				contentType: "application/json",
				body:        `{"login":"testuser","password":"testPass123"`,
			},
			userService: func(mc *minimock.Controller) UserService {
				return mocks.NewUserServiceMock(mc)
			},
			authService: func(mc *minimock.Controller) AuthService {
				return mocks.NewAuthServiceMock(mc)
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "invalid password",
			request: request{
				contentType: "application/json",
				body:        `{"login":"testuser@email.com","password":"testpass"`,
			},
			userService: func(mc *minimock.Controller) UserService {
				return mocks.NewUserServiceMock(mc)
			},
			authService: func(mc *minimock.Controller) AuthService {
				return mocks.NewAuthServiceMock(mc)
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "empty body",
			request: request{
				contentType: "application/json",
				body:        ``,
			},
			userService: func(mc *minimock.Controller) UserService {
				return mocks.NewUserServiceMock(mc)
			},
			authService: func(mc *minimock.Controller) AuthService {
				return mocks.NewAuthServiceMock(mc)
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "user already exists",
			request: request{
				contentType: "application/json",
				body:        `{"login":"existinguser@email.com","password":"testPass123"}`,
			},
			userService: func(mc *minimock.Controller) UserService {
				return mocks.NewUserServiceMock(mc).
					RegisterMock.
					Expect(context.Background(), "existinguser@email.com", "testPass123").
					Return(int64(0), models.ErrUserAlreadyExists)
			},
			authService: func(mc *minimock.Controller) AuthService {
				return mocks.NewAuthServiceMock(mc)
			},
			expected: http.StatusConflict,
		},
		{
			name: "internal server error",
			request: request{
				contentType: "application/json",
				body:        `{"login":"newuser@email.com","password":"testPass123"}`,
			},
			userService: func(mc *minimock.Controller) UserService {
				return mocks.NewUserServiceMock(mc).
					RegisterMock.
					Expect(context.Background(), "newuser@email.com", "testPass123").
					Return(int64(0), errors.New("database error"))
			},
			authService: func(mc *minimock.Controller) AuthService {
				return mocks.NewAuthServiceMock(mc)
			},
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			defer mc.Wait(time.Second)

			handler := NewHandler(
				tt.userService(mc),
				tt.authService(mc),
				zerolog.Nop(),
			)

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(tt.request.body))
			req.Header.Set("Content-Type", tt.request.contentType)

			rec := httptest.NewRecorder()

			handler.Handle(rec, req)

			assert.Equal(t, tt.expected, rec.Code)
		})
	}
}
