package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/hydra13/gophermart/internal/models"
	"github.com/hydra13/gophermart/internal/services/worker/mocks"
)

func TestWorker_fetchAndDistributeOrders(t *testing.T) {
	tests := []struct {
		name         string
		client       func(mc *minimock.Controller) *mocks.AccualClientMock
		orderService func(mc *minimock.Controller) *mocks.OrderServiceMock
		wantErr      bool
	}{
		{
			name: "success",
			client: func(mc *minimock.Controller) *mocks.AccualClientMock {
				return mocks.NewAccualClientMock(mc)
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
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "error getting orders",
			client: func(mc *minimock.Controller) *mocks.AccualClientMock {
				return mocks.NewAccualClientMock(mc)
			},
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					GetOrdersForCheckingStatusMock.
					Expect(minimock.AnyContext).
					Return(nil, errors.New("database error"))
			},
			wantErr: true,
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

			ordersChan := make(chan models.Order, numWorkers)
			gotErr := w.fetchAndDistributeOrders(context.Background(), ordersChan)
			close(ordersChan)

			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			assert.NoError(t, gotErr)
		})
	}
}

func TestWorker_worker(t *testing.T) {
	tests := []struct {
		name         string
		client       func(mc *minimock.Controller) *mocks.AccualClientMock
		orderService func(mc *minimock.Controller) *mocks.OrderServiceMock
		sendOrder    bool
		order        models.Order
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
					UpdateOrderMock.
					Expect(minimock.AnyContext, models.Order{
						Number:  "4440",
						Status:  models.OrderStatusProcessed,
						Accrual: 100,
						UserID:  777,
					}).
					Return(nil)
			},
			sendOrder: true,
			order: models.Order{
				Number: "4440",
				Status: models.OrderStatusProcessing,
				UserID: 777,
			},
		},
		{
			name: "error getting order status",
			client: func(mc *minimock.Controller) *mocks.AccualClientMock {
				return mocks.NewAccualClientMock(mc).
					GetOrderStatusMock.
					Expect(minimock.AnyContext, "5555").
					Return(models.AccrualResponse{}, errors.New("network error"))
			},
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc)
			},
			sendOrder: true,
			order: models.Order{
				Number: "5555",
				Status: models.OrderStatusProcessing,
				UserID: 888,
			},
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

			ordersChan := make(chan models.Order, 1)

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// Send order to worker if needed
			if tt.sendOrder {
				ordersChan <- tt.order
			}

			// Close channel to signal worker to finish
			close(ordersChan)

			// Run worker
			w.worker(ctx, ordersChan, 0)
		})
	}
}
