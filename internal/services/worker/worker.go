//go:generate minimock -i .AccualClient,.OrderService -o mocks -s _mock.go -g
package worker

import (
	"context"
	"time"

	"github.com/hydra13/gophermart/internal/models"
	"github.com/rs/zerolog"
)

type AccualClient interface {
	GetOrderStatus(ctx context.Context, orderNumber string) (models.AccrualResponse, error)
}

type OrderService interface {
	GetOrdersForCheckingStatus(ctx context.Context) ([]string, error)
	UpdateOrder(ctx context.Context, orderNumber string, status string, accrual int64) error
}

const sleepTimeout = 3 * time.Second

type Worker struct {
	client       AccualClient
	orderService OrderService
	log          zerolog.Logger
}

func NewWorker(
	client AccualClient,
	orderService OrderService,
	log zerolog.Logger,
) *Worker {
	return &Worker{
		client:       client,
		orderService: orderService,
		log:          log,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(sleepTimeout):
			err := w.doWork(ctx)

			if err != nil {
				return err
			}
		}
	}
}

func (w *Worker) doWork(ctx context.Context) error {
	err := w.handleProcessingOrders(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (w *Worker) handleProcessingOrders(ctx context.Context) error {
	orders, err := w.orderService.GetOrdersForCheckingStatus(ctx)
	orders = append(orders, "4440", "4444", "4444444444444448")
	if err != nil {
		return err
	}

	for _, orderNumber := range orders {
		resp, err := w.client.GetOrderStatus(ctx, orderNumber)
		if err != nil {
			w.log.Error().
				Err(err).
				Str("order_number", orderNumber).
				Msg("error get order status")

			return err
		}

		w.log.Debug().
			Str("order_number", resp.Order).
			Str("status", resp.Status).
			Int64("accrual", resp.Accrual).
			Msg("got order status")

		if resp.Status == models.OrderStatusInvalid || resp.Status == models.OrderStatusProcessed {
			err := w.orderService.UpdateOrder(ctx, orderNumber, resp.Status, resp.Accrual)
			if err != nil {
				w.log.Error().
					Err(err).
					Str("order_number", orderNumber).
					Msg("error update order status")

				return err
			}
		}
	}

	return nil
}
