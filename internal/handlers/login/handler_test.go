package loginhandler

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

	"github.com/hydra13/gophermart/internal/handlers/login/mocks"
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
		userService func(mc *minimock.Controller) *mocks.UserServiceMock
		authService func(mc *minimock.Controller) *mocks.AuthServiceMock
		expected    int
	}{
		{
			name: "success",
			request: request{
				contentType: "application/json",
				body:        `{"login":"testuser","password":"testpass"}`,
			},
			userService: func(mc *minimock.Controller) *mocks.UserServiceMock {
				return mocks.NewUserServiceMock(mc).
					LoginMock.
					Expect(context.Background(), "testuser", "testpass").
					Return(int64(1234567890), nil)
			},
			authService: func(mc *minimock.Controller) *mocks.AuthServiceMock {
				return mocks.NewAuthServiceMock(mc).
					SetAuthCookieMock.
					ExpectUserIDParam2(int64(1234567890)).
					Return()
			},
			expected: http.StatusOK,
		},
		{
			name: "invalid content-type",
			request: request{
				contentType: "text/html",
				body:        `{"login":"testuser","password":"testpass"}`,
			},
			userService: func(mc *minimock.Controller) *mocks.UserServiceMock {
				return mocks.NewUserServiceMock(mc)
			},
			authService: func(mc *minimock.Controller) *mocks.AuthServiceMock {
				return mocks.NewAuthServiceMock(mc)
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "invalid json",
			request: request{
				contentType: "application/json",
				body:        `{"login":"testuser","password":"testpass"`,
			},
			userService: func(mc *minimock.Controller) *mocks.UserServiceMock {
				return mocks.NewUserServiceMock(mc)
			},
			authService: func(mc *minimock.Controller) *mocks.AuthServiceMock {
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
			userService: func(mc *minimock.Controller) *mocks.UserServiceMock {
				return mocks.NewUserServiceMock(mc)
			},
			authService: func(mc *minimock.Controller) *mocks.AuthServiceMock {
				return mocks.NewAuthServiceMock(mc)
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "user not found",
			request: request{
				contentType: "application/json",
				body:        `{"login":"nonexistentuser","password":"testpass"}`,
			},
			userService: func(mc *minimock.Controller) *mocks.UserServiceMock {
				return mocks.NewUserServiceMock(mc).
					LoginMock.
					Expect(context.Background(), "nonexistentuser", "testpass").
					Return(int64(0), models.ErrUserNotFound)
			},
			authService: func(mc *minimock.Controller) *mocks.AuthServiceMock {
				return mocks.NewAuthServiceMock(mc)
			},
			expected: http.StatusUnauthorized,
		},
		{
			name: "internal server error",
			request: request{
				contentType: "application/json",
				body:        `{"login":"newuser","password":"testpass"}`,
			},
			userService: func(mc *minimock.Controller) *mocks.UserServiceMock {
				return mocks.NewUserServiceMock(mc).
					LoginMock.
					Expect(context.Background(), "newuser", "testpass").
					Return(int64(0), errors.New("database error"))
			},
			authService: func(mc *minimock.Controller) *mocks.AuthServiceMock {
				return mocks.NewAuthServiceMock(mc)
			},
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)
			defer mc.Wait(time.Second)

			handler := NewHandler(
				tt.userService(mc),
				tt.authService(mc),
				zerolog.Nop(),
			)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tt.request.body))
			req.Header.Set("Content-Type", tt.request.contentType)

			rec := httptest.NewRecorder()

			handler.Handle(rec, req)

			assert.Equal(t, tt.expected, rec.Code)
		})
	}
}
