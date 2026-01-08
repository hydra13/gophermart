package withdrawal

import (
	"context"

	"github.com/hydra13/gophermart/internal/models"
)

type WithdrawalRepository interface {
	GetByUserID(ctx context.Context, userID int64) ([]models.Withdrawal, error)
}

type WithdrawalService struct {
	repo WithdrawalRepository
}

func NewWithdrawalService(repo WithdrawalRepository) *WithdrawalService {
	return &WithdrawalService{repo: repo}
}

func (w *WithdrawalService) GetWithdrawalsByUser(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	res, err := w.repo.GetByUserID(ctx, userID)
	if err != nil {
		return []models.Withdrawal{}, err
	}
	return res, nil
}
