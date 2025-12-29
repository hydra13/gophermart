package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/hydra13/gophermart/internal/models"
	"github.com/hydra13/gophermart/internal/services/worker/mocks"
)

func TestWorker_doWork(t *testing.T) {
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
		{
			name: "error handled gracefully",
			client: func(mc *minimock.Controller) *mocks.AccualClientMock {
				return mocks.NewAccualClientMock(mc).
					GetOrderStatusMock.
					Expect(minimock.AnyContext, "5555").
					Return(models.AccrualResponse{}, errors.New("network error"))
			},
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					GetOrdersForCheckingStatusMock.
					Expect(minimock.AnyContext).
					Return([]models.Order{
						{
							Number: "5555",
							Status: models.OrderStatusProcessing,
							UserID: 888,
						},
					}, nil)
			},
			wantErr: true, // doWork should return the error from handleProcessingOrders
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
			gotErr := w.doWork(context.Background())
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			assert.NoError(t, gotErr)
		})
	}
}

func TestWorker_handleProcessingOrders(t *testing.T) {
	tests := []struct {
		name         string
		client       func(mc *minimock.Controller) *mocks.AccualClientMock
		orderService func(mc *minimock.Controller) *mocks.OrderServiceMock
		wantErr      bool
	}{
		{
			name: "success with processed order",
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
		{
			name: "success with invalid order",
			client: func(mc *minimock.Controller) *mocks.AccualClientMock {
				return mocks.NewAccualClientMock(mc).
					GetOrderStatusMock.
					Expect(minimock.AnyContext, "5555").
					Return(models.AccrualResponse{
						Order:   "5555",
						Status:  models.OrderStatusInvalid,
						Accrual: 0,
					}, nil)
			},
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					GetOrdersForCheckingStatusMock.
					Expect(minimock.AnyContext).
					Return([]models.Order{
						{
							Number: "5555",
							Status: models.OrderStatusProcessing,
							UserID: 888,
						},
					}, nil).
					UpdateOrderMock.
					Expect(minimock.AnyContext, models.Order{
						Number:  "5555",
						Status:  models.OrderStatusInvalid,
						Accrual: 0,
						UserID:  888,
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success with processing order - no update",
			client: func(mc *minimock.Controller) *mocks.AccualClientMock {
				return mocks.NewAccualClientMock(mc).
					GetOrderStatusMock.
					Expect(minimock.AnyContext, "6666").
					Return(models.AccrualResponse{
						Order:   "6666",
						Status:  models.OrderStatusProcessing,
						Accrual: 0,
					}, nil)
			},
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					GetOrdersForCheckingStatusMock.
					Expect(minimock.AnyContext).
					Return([]models.Order{
						{
							Number: "6666",
							Status: models.OrderStatusProcessing,
							UserID: 999,
						},
					}, nil)
					// No UpdateOrderMock expectation since status is still processing
			},
			wantErr: false,
		},
		{
			name: "multiple orders processing",
			client: func(mc *minimock.Controller) *mocks.AccualClientMock {
				clientMock := mocks.NewAccualClientMock(mc)
				clientMock.GetOrderStatusMock.When(minimock.AnyContext, "1111").Then(models.AccrualResponse{
					Order:   "1111",
					Status:  models.OrderStatusProcessed,
					Accrual: 150,
				}, nil)
				clientMock.GetOrderStatusMock.When(minimock.AnyContext, "2222").Then(models.AccrualResponse{
					Order:   "2222",
					Status:  models.OrderStatusInvalid,
					Accrual: 0,
				}, nil)
				return clientMock
			},
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				orderServiceMock := mocks.NewOrderServiceMock(mc)
				orderServiceMock.GetOrdersForCheckingStatusMock.Expect(minimock.AnyContext).Return([]models.Order{
					{
						Number: "1111",
						Status: models.OrderStatusProcessing,
						UserID: 111,
					},
					{
						Number: "2222",
						Status: models.OrderStatusProcessing,
						UserID: 222,
					},
				}, nil)
				orderServiceMock.UpdateOrderMock.When(minimock.AnyContext, models.Order{
					Number:  "1111",
					Status:  models.OrderStatusProcessed,
					Accrual: 150,
					UserID:  111,
				}).Then(nil)
				orderServiceMock.UpdateOrderMock.When(minimock.AnyContext, models.Order{
					Number:  "2222",
					Status:  models.OrderStatusInvalid,
					Accrual: 0,
					UserID:  222,
				}).Then(nil)
				return orderServiceMock
			},
			wantErr: false,
		},
		{
			name: "error getting orders for checking",
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
		{
			name: "error getting order status",
			client: func(mc *minimock.Controller) *mocks.AccualClientMock {
				return mocks.NewAccualClientMock(mc).
					GetOrderStatusMock.
					Expect(minimock.AnyContext, "7777").
					Return(models.AccrualResponse{}, errors.New("network error"))
			},
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					GetOrdersForCheckingStatusMock.
					Expect(minimock.AnyContext).
					Return([]models.Order{
						{
							Number: "7777",
							Status: models.OrderStatusProcessing,
							UserID: 333,
						},
					}, nil)
			},
			wantErr: true,
		},
		{
			name: "error updating order",
			client: func(mc *minimock.Controller) *mocks.AccualClientMock {
				return mocks.NewAccualClientMock(mc).
					GetOrderStatusMock.
					Expect(minimock.AnyContext, "8888").
					Return(models.AccrualResponse{
						Order:   "8888",
						Status:  models.OrderStatusProcessed,
						Accrual: 200,
					}, nil)
			},
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					GetOrdersForCheckingStatusMock.
					Expect(minimock.AnyContext).
					Return([]models.Order{
						{
							Number: "8888",
							Status: models.OrderStatusProcessing,
							UserID: 444,
						},
					}, nil).
					UpdateOrderMock.
					Expect(minimock.AnyContext, models.Order{
						Number:  "8888",
						Status:  models.OrderStatusProcessed,
						Accrual: 200,
						UserID:  444,
					}).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "no orders to process",
			client: func(mc *minimock.Controller) *mocks.AccualClientMock {
				return mocks.NewAccualClientMock(mc)
			},
			orderService: func(mc *minimock.Controller) *mocks.OrderServiceMock {
				return mocks.NewOrderServiceMock(mc).
					GetOrdersForCheckingStatusMock.
					Expect(minimock.AnyContext).
					Return([]models.Order{}, nil)
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