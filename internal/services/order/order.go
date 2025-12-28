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
		return err
	}

	return nil
}
