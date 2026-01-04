//go:generate minimock -i .AccualClient,.OrderService -o mocks -s _mock.go -g
package worker

import (
	"context"
	"sync"
	"time"

	"github.com/hydra13/gophermart/internal/models"
	"github.com/rs/zerolog"
)

type AccualClient interface {
	GetOrderStatus(ctx context.Context, orderNumber string) (models.AccrualResponse, error)
}

type OrderService interface {
	GetOrdersForCheckingStatus(ctx context.Context) ([]models.Order, error)
	UpdateOrder(ctx context.Context, order models.Order) error
}

const (
	sleepTimeout = 10 * time.Second
	numWorkers   = 3
)

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
	ordersChan := make(chan models.Order, numWorkers)

	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()

	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for workerID := range numWorkers {
		go func() {
			defer wg.Done()

			w.worker(workerCtx, ordersChan, workerID)
		}()
	}

	for {
		select {
		case <-ctx.Done():
			close(ordersChan)
			wg.Wait()
			return nil
		case <-time.After(sleepTimeout):
			err := w.fetchAndDistributeOrders(workerCtx, ordersChan)
			if err != nil {
				w.log.Error().
					Err(err).
					Msg("worker: fetching orders error")

				cancelWorkers()
				close(ordersChan)
				wg.Wait()
				return err
			}
		}
	}
}

func (w *Worker) fetchAndDistributeOrders(ctx context.Context, ordersChan chan<- models.Order) error {
	orders, err := w.orderService.GetOrdersForCheckingStatus(ctx)
	if err != nil {
		return err
	}

	for _, order := range orders {
		select {
		case <-ctx.Done():
			return nil
		case ordersChan <- order:
		}
	}

	return nil
}

func (w *Worker) worker(ctx context.Context, ordersChan <-chan models.Order, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		case order, ok := <-ordersChan:
			if !ok {
				return
			}

			resp, err := w.client.GetOrderStatus(ctx, order.Number)
			if err != nil {
				w.log.Error().
					Err(err).
					Str("order_number", order.Number).
					Int("worker_id", workerID).
					Msg("worker: error get order status")

				return
			}

			w.log.Debug().
				Str("order_number", resp.Order).
				Str("status", resp.Status).
				Int64("accrual", resp.Accrual).
				Int("worker_id", workerID).
				Msg("worker: got order status")

			if resp.Status == models.OrderStatusInvalid || resp.Status == models.OrderStatusProcessed {
				order.Status = resp.Status
				order.Accrual = resp.Accrual

				err := w.orderService.UpdateOrder(ctx, order)
				if err != nil {
					w.log.Error().
						Err(err).
						Str("order_number", order.Number).
						Int("worker_id", workerID).
						Msg("error update order status")

					return
				}
			}
		}
	}
}
