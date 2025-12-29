package order

import (
	"context"

	"github.com/hydra13/gophermart/internal/models"
	repositories "github.com/hydra13/gophermart/internal/repositories/order"
)

type OrderRepository interface {
	CheckAndCreate(ctx context.Context, number string, userID int64) error
	GetByUserID(ctx context.Context, userID int64) ([]models.Order, error)
	GetProcessingOrders(ctx context.Context) ([]models.Order, error)
	UpdateStatus(ctx context.Context, orderNumber string, status string, accrual int64) error
}

type TransactionRepository interface {
	UpdateOrderAndAccount(ctx context.Context, order models.Order) error
}

type OrderService struct {
	orderRepo OrderRepository
	txRepo    TransactionRepository
}

func NewOrderService(
	orderRepo OrderRepository,
	txRepo TransactionRepository,
) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		txRepo:    txRepo,
	}
}

func (s *OrderService) AddOrder(ctx context.Context, orderNumber string, userID int64) error {
	err := s.orderRepo.CheckAndCreate(ctx, orderNumber, userID)
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

func (s *OrderService) GetOrdersByUser(ctx context.Context, userID int64) ([]models.Order, error) {
	orders, err := s.orderRepo.GetByUserID(ctx, userID)
	if err != nil {
		// TODO: надо мапить на ошибки сервисного слоя а не прокидывать ошибки репозитория
		return nil, err
	}

	return orders, nil
}

func (s *OrderService) GetOrdersForCheckingStatus(ctx context.Context) ([]models.Order, error) {
	orders, err := s.orderRepo.GetProcessingOrders(ctx)
	if err != nil {
		// TODO: надо мапить на ошибки сервисного слоя а не прокидывать ошибки репозитория
		return nil, err
	}

	return orders, nil
}

func (s *OrderService) UpdateOrder(
	ctx context.Context,
	order models.Order,
) error {
	err := s.txRepo.UpdateOrderAndAccount(ctx, order)
	if err != nil {
		// TODO: надо мапить на ошибки сервисного слоя а не прокидывать ошибки репозитория
		return err
	}

	return nil
}
