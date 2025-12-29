package order

import (
	"context"

	"github.com/hydra13/gophermart/internal/models"
	repositories "github.com/hydra13/gophermart/internal/repositories/order"
)

type OrderService struct {
	repo *repositories.OrderRepository
}

func NewOrderService(repo *repositories.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) AddOrder(ctx context.Context, orderNumber string, userID int64) error {
	err := s.repo.CheckAndCreate(ctx, orderNumber, userID)
	if err == repositories.ErrConflict {
		return models.ErrConflict
	}

	if err == repositories.ErrOrderExists {
		return models.ErrOrderAlreadyExists
	}

	if err != nil {
		// TODO: надо мапить на ошибки сервисного слоя а не прокидывать ошибки репозитория
		return err
	}

	return nil
}

func (s *OrderService) GetOrdersForCheckingStatus(ctx context.Context) ([]string, error) {
	orders, err := s.repo.GetProcessingOrders(ctx)
	if err != nil {
		// TODO: надо мапить на ошибки сервисного слоя а не прокидывать ошибки репозитория
		return nil, err
	}

	return orders, nil
}

func (s *OrderService) UpdateOrder(
	ctx context.Context,
	orderNumber string,
	status string,
	accrual int64,
) error {
	err := s.repo.UpdateStatus(ctx, orderNumber, status, accrual)
	if err != nil {
		// TODO: надо мапить на ошибки сервисного слоя а не прокидывать ошибки репозитория
		return err
	}

	return nil
}
