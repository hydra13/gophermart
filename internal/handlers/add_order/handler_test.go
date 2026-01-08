package addorderhandler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hydra13/gophermart/internal/handlers/add_order/mocks"
	"github.com/hydra13/gophermart/internal/models"
	authContext "github.com/hydra13/gophermart/internal/services/auth_context"
)

type request struct {
	contentType string
	body        string
}

func TestHandler_Handle(t *testing.T) {
	tests := []struct {
		name         string
		request      request
		userID       int64
		orderService func(mc *minimock.Controller) *mocks.OrderServiceMock
		expected     int
	}{
		{
			name: "success",
			request: request{
				contentType: "text/plain",
				body:        "4440",
			},
			userID: 777,
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					AddOrderMock.
					Expect(minimock.AnyContext, "4440", 777).
					Return(nil)
			},
			expected: http.StatusAccepted,
		},
		{
			name: "invalid content-type",
			request: request{
				contentType: "application/json",
				body:        "4440",
			},
			userID: 777,
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc)
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "invalid order number",
			request: request{
				contentType: "text/plain",
				body:        "1234",
			},
			userID: 777,
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc)
			},
			expected: http.StatusUnprocessableEntity,
		},
		{
			name: "empty body",
			request: request{
				contentType: "text/plain",
				body:        "",
			},
			userID: 777,
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc)
			},
			expected: http.StatusBadRequest,
		},
		{
			name: "order already exists",
			request: request{
				contentType: "text/plain",
				body:        "4440",
			},
			userID: 777,
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					AddOrderMock.
					Expect(minimock.AnyContext, "4440", 777).
					Return(models.ErrOrderAlreadyExists)
			},
			expected: http.StatusOK,
		},
		{
			name: "order already added by another user",
			request: request{
				contentType: "text/plain",
				body:        "4440",
			},
			userID: 777,
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					AddOrderMock.
					Expect(minimock.AnyContext, "4440", 777).
					Return(models.ErrConflict)
			},
			expected: http.StatusConflict,
		},
		{
			name: "internal server error",
			request: request{
				contentType: "text/plain",
				body:        "4440",
			},
			userID: 777,
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					AddOrderMock.
					Expect(minimock.AnyContext, "4440", 777).
					Return(errors.New("database error"))
			},
			expected: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)

			handler := NewHandler(
				tt.orderService(mc),
				zerolog.Nop(),
			)

			ctx := authContext.CreateContextWithUserID(context.Background(), tt.userID)

			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/orders",
				strings.NewReader(tt.request.body),
			)
			require.NoError(t, err)
			req.Header.Set("Content-Type", tt.request.contentType)

			rec := httptest.NewRecorder()

			handler.Handle(rec, req)

			assert.Equal(t, tt.expected, rec.Code)
		})
	}
}
