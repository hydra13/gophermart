package withdrawals

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hydra13/gophermart/internal/handlers/withdrawals/mocks"
	"github.com/hydra13/gophermart/internal/models"
	authContext "github.com/hydra13/gophermart/internal/services/auth_context"
	"github.com/hydra13/gophermart/internal/utils"
)

func TestHandler_Handle(t *testing.T) {
	tests := []struct {
		name              string
		userID            int64
		withdrawalService func(mc *minimock.Controller) *mocks.WithdrawalServiceMock
		expectedCode      int
		expectedBody      string
	}{
		{
			name:   "success with withdrawals",
			userID: 123,
			withdrawalService: func(mc *minimock.Controller) *mocks.WithdrawalServiceMock {
				processedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
				withdrawals := []models.Withdrawal{
					{
						OrderNumber: "12345",
						Sum:         100_00, // 100.00 in cents
						ProcessedAt: processedAt,
						UserID:      123,
					},
					{
						OrderNumber: "67890",
						Sum:         50_50, // 50.50 in cents
						ProcessedAt: processedAt.Add(time.Hour),
						UserID:      123,
					},
				}
				return mocks.NewWithdrawalServiceMock(mc).
					GetWithdrawalsByUserMock.
					Expect(minimock.AnyContext, int64(123)).
					Return(withdrawals, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `[
				{"order":"12345","sum":100,"processed_at":"2023-01-01T12:00:00Z"},
				{"order":"67890","sum":50.5,"processed_at":"2023-01-01T13:00:00Z"}
			]`,
		},
		{
			name:   "no withdrawals",
			userID: 456,
			withdrawalService: func(mc *minimock.Controller) *mocks.WithdrawalServiceMock {
				return mocks.NewWithdrawalServiceMock(mc).
					GetWithdrawalsByUserMock.
					Expect(minimock.AnyContext, int64(456)).
					Return(nil, models.ErrNoWithdrawals)
			},
			expectedCode: http.StatusNoContent,
			expectedBody: ``,
		},
		{
			name:   "service error",
			userID: 789,
			withdrawalService: func(mc *minimock.Controller) *mocks.WithdrawalServiceMock {
				return mocks.NewWithdrawalServiceMock(mc).
					GetWithdrawalsByUserMock.
					Expect(minimock.AnyContext, int64(789)).
					Return(nil, errors.New("database error"))
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
				tt.withdrawalService(mc),
				zerolog.Nop(),
			)

			ctx := authContext.CreateContextWithUserID(context.Background(), tt.userID)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/withdrawals", nil)
			require.NoError(t, err)

			rec := httptest.NewRecorder()
			handler.Handle(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedBody != "" {
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

				var expected []ResponseRecord
				err = json.Unmarshal([]byte(tt.expectedBody), &expected)
				require.NoError(t, err)

				var actual []ResponseRecord
				err = json.Unmarshal(rec.Body.Bytes(), &actual)
				require.NoError(t, err)

				assert.Equal(t, expected, actual)
			}
		})
	}
}

func TestHandler_toResponse(t *testing.T) {
	processedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	withdrawals := []models.Withdrawal{
		{
			OrderNumber: "12345",
			Sum:         150_75, // 150.75 in cents
			ProcessedAt: processedAt,
			UserID:      123,
		},
		{
			OrderNumber: "67890",
			Sum:         0, // No sum
			ProcessedAt: processedAt.Add(time.Hour),
			UserID:      123,
		},
	}

	handler := &Handler{}
	response := handler.toResponse(withdrawals)

	expected := []ResponseRecord{
		{
			Order:       "12345",
			Sum:         utils.ToRub(150_75), // 150.75
			ProcessedAt: processedAt,
		},
		{
			Order:       "67890",
			Sum:         0,
			ProcessedAt: processedAt.Add(time.Hour),
		},
	}

	assert.Equal(t, expected, response)
}