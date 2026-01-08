package withdrawhandler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hydra13/gophermart/internal/handlers/withdraw/mocks"
	"github.com/hydra13/gophermart/internal/models"
	authContext "github.com/hydra13/gophermart/internal/services/auth_context"
	"github.com/hydra13/gophermart/internal/utils"
)

type request struct {
	contentType string
	body        string
}

func TestHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		request        request
		accountService func(mc *minimock.Controller) *mocks.AccountServiceMock
		expectedCode   int
	}{
		{
			name:   "success",
			userID: 123,
			request: request{
				contentType: "application/json",
				body:        `{"order":"4440","sum":100.50}`,
			},
			accountService: func(mc *minimock.Controller) *mocks.AccountServiceMock {
				return mocks.NewAccountServiceMock(mc).
					WithdrawMock.
					Expect(minimock.AnyContext, models.Withdrawal{
						OrderNumber: "4440",
						Sum:         utils.FromRub(100.50),
						UserID:      123,
					}).
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:   "invalid content-type",
			userID: 123,
			request: request{
				contentType: "text/plain",
				body:        `{"order":"4440","sum":100.50}`,
			},
			accountService: func(mc *minimock.Controller) *mocks.AccountServiceMock {
				return mocks.NewAccountServiceMock(mc)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:   "invalid json",
			userID: 123,
			request: request{
				contentType: "application/json",
				body:        `{"order":"4440","sum":100.50`,
			},
			accountService: func(mc *minimock.Controller) *mocks.AccountServiceMock {
				return mocks.NewAccountServiceMock(mc)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:   "invalid order number",
			userID: 123,
			request: request{
				contentType: "application/json",
				body:        `{"order":"1234","sum":100.50}`,
			},
			accountService: func(mc *minimock.Controller) *mocks.AccountServiceMock {
				return mocks.NewAccountServiceMock(mc)
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:   "invalid withdraw amount - negative",
			userID: 123,
			request: request{
				contentType: "application/json",
				body:        `{"order":"4440","sum":-50.00}`,
			},
			accountService: func(mc *minimock.Controller) *mocks.AccountServiceMock {
				return mocks.NewAccountServiceMock(mc)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:   "invalid withdraw amount - zero",
			userID: 123,
			request: request{
				contentType: "application/json",
				body:        `{"order":"4440","sum":0}`,
			},
			accountService: func(mc *minimock.Controller) *mocks.AccountServiceMock {
				return mocks.NewAccountServiceMock(mc)
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:   "not enough money",
			userID: 123,
			request: request{
				contentType: "application/json",
				body:        `{"order":"4440","sum":1000.00}`,
			},
			accountService: func(mc *minimock.Controller) *mocks.AccountServiceMock {
				return mocks.NewAccountServiceMock(mc).
					WithdrawMock.
					Expect(minimock.AnyContext, models.Withdrawal{
						OrderNumber: "4440",
						Sum:         utils.FromRub(1000.00),
						UserID:      123,
					}).
					Return(models.ErrNotEnoughMoney)
			},
			expectedCode: http.StatusPaymentRequired,
		},
		{
			name:   "service error",
			userID: 123,
			request: request{
				contentType: "application/json",
				body:        `{"order":"4440","sum":100.50}`,
			},
			accountService: func(mc *minimock.Controller) *mocks.AccountServiceMock {
				return mocks.NewAccountServiceMock(mc).
					WithdrawMock.
					Expect(minimock.AnyContext, models.Withdrawal{
						OrderNumber: "4440",
						Sum:         utils.FromRub(100.50),
						UserID:      123,
					}).
					Return(errors.New("database error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:   "empty body",
			userID: 123,
			request: request{
				contentType: "application/json",
				body:        ``,
			},
			accountService: func(mc *minimock.Controller) *mocks.AccountServiceMock {
				return mocks.NewAccountServiceMock(mc)
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)

			handler := NewHandler(
				tt.accountService(mc),
				zerolog.Nop(),
			)

			ctx := authContext.CreateContextWithUserID(context.Background(), tt.userID)
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/withdraw", bytes.NewBufferString(tt.request.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", tt.request.contentType)

			rec := httptest.NewRecorder()
			handler.Handle(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}
