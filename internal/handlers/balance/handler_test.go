package balancehandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hydra13/gophermart/internal/handlers/balance/mocks"
	"github.com/hydra13/gophermart/internal/models"
	authContext "github.com/hydra13/gophermart/internal/services/auth_context"
)

func TestHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		accountService func(mc *minimock.Controller) *mocks.AccountServiceMock
		expectedCode   int
		expectedBody   string
	}{
		{
			name:   "success with balance",
			userID: 123,
			accountService: func(mc *minimock.Controller) *mocks.AccountServiceMock {
				account := models.Account{
					UserID:    123,
					Current:   150_75,
					Withdrawn: 50_25,
				}
				return mocks.NewAccountServiceMock(mc).
					GetAccountMock.
					Expect(minimock.AnyContext, int64(123)).
					Return(account, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"current":150.75,"withdrawn":50.25}`,
		},
		{
			name:   "success with zero balance",
			userID: 456,
			accountService: func(mc *minimock.Controller) *mocks.AccountServiceMock {
				account := models.Account{
					UserID:    456,
					Current:   0,
					Withdrawn: 0,
				}
				return mocks.NewAccountServiceMock(mc).
					GetAccountMock.
					Expect(minimock.AnyContext, int64(456)).
					Return(account, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"current":0,"withdrawn":0}`,
		},
		{
			name:   "service error",
			userID: 789,
			accountService: func(mc *minimock.Controller) *mocks.AccountServiceMock {
				return mocks.NewAccountServiceMock(mc).
					GetAccountMock.
					Expect(minimock.AnyContext, int64(789)).
					Return(models.Account{}, errors.New("database error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: ``,
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
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/balance", nil)
			require.NoError(t, err)

			rec := httptest.NewRecorder()
			handler.Handle(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedBody != "" {
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

				var expected JSONResponse
				err = json.Unmarshal([]byte(tt.expectedBody), &expected)
				require.NoError(t, err)

				var actual JSONResponse
				err = json.Unmarshal(rec.Body.Bytes(), &actual)
				require.NoError(t, err)

				assert.Equal(t, expected, actual)
			}
		})
	}
}
