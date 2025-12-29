package worker

import (
	"context"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/hydra13/gophermart/internal/models"
	"github.com/hydra13/gophermart/internal/services/worker/mocks"
)

func TestWorker_handleProcessingOrders(t *testing.T) {
	tests := []struct {
		name         string
		client       func(mc *minimock.Controller) *mocks.AccualClientMock
		orderService func(mc *minimock.Controller) *mocks.OrderServiceMock
		wantErr      bool
	}{
		{
			name: "success",
			client: func(mc *minimock.Controller) *mocks.AccualClientMock {
				return mocks.NewAccualClientMock(mc).
					GetOrderStatusMock.
					Expect(minimock.AnyContext, "4440").
					Return(models.AccrualResponse{
						Order:   "4440",
						Status:  models.OrderStatusProcessed,
						Accrual: 100,
					}, nil)
			},
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					GetOrdersForCheckingStatusMock.
					Expect(minimock.AnyContext).
					Return([]models.Order{
						{
							Number: "4440",
							Status: models.OrderStatusProcessing,
							UserID: 777,
						},
					}, nil).
					UpdateOrderMock.
					Expect(minimock.AnyContext, models.Order{
						Number:  "4440",
						Status:  models.OrderStatusProcessed,
						Accrual: 100,
						UserID:  777,
					}).
					Return(nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)

			w := NewWorker(
				tt.client(mc),
				tt.orderService(mc),
				zerolog.Nop(),
			)
			gotErr := w.handleProcessingOrders(context.Background())
			if tt.wantErr {
				assert.Error(t, gotErr)

				return
			}

			assert.NoError(t, gotErr)
		})
	}
}
