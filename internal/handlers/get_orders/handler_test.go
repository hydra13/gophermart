package getorderhandler

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

	"github.com/hydra13/gophermart/internal/handlers/get_orders/mocks"
	"github.com/hydra13/gophermart/internal/models"
	authContext "github.com/hydra13/gophermart/internal/services/auth_context"
)

func TestHandler_Handle(t *testing.T) {
	tests := []struct {
		name         string
		userID       int64
		orderService func(mc *minimock.Controller) *mocks.OrderServiceMock
		expectedCode int
		expectedBody string
	}{
		{
			name:   "success with orders",
			userID: 123,
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				uploadedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
				orders := []models.Order{
					{
						Number:     "12345",
						Status:     "NEW",
						Accrual:    100_00,
						UploadedAt: uploadedAt,
					},
					{
						Number:     "67890",
						Status:     "PROCESSED",
						Accrual:    500_00,
						UploadedAt: uploadedAt.Add(time.Hour),
					},
				}
				return mocks.NewOrderServiceMock(mc).
					GetOrdersByUserMock.
					Expect(minimock.AnyContext, int64(123)).
					Return(orders, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `[
				{"number":"12345","status":"NEW","accrual":100,"upload_at":"2023-01-01T12:00:00Z"},
				{"number":"67890","status":"PROCESSED","accrual":500,"upload_at":"2023-01-01T13:00:00Z"}
			]`,
		},
		{
			name:   "no orders",
			userID: 456,
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					GetOrdersByUserMock.
					Expect(minimock.AnyContext, int64(456)).
					Return([]models.Order{}, nil)
			},
			expectedCode: http.StatusNoContent,
			expectedBody: ``,
		},
		{
			name:   "service error",
			userID: 789,
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					GetOrdersByUserMock.
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
				tt.orderService(mc),
				zerolog.Nop(),
			)

			ctx := authContext.CreateContextWithUserID(context.Background(), tt.userID)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/orders", nil)
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
