package account

import (
	"context"
	"errors"

	"github.com/hydra13/gophermart/internal/models"
	"github.com/hydra13/gophermart/internal/repositories"
)

type AccountRepository interface {
	Get(ctx context.Context, userID int64) (models.Account, error)
}

type TransactionRepository interface {
	Withdraw(ctx context.Context, withdrawal models.Withdrawal) error
}

type AccountService struct {
	accountRepository     AccountRepository
	transactionRepository TransactionRepository
}

func NewAccountService(
	accountRepository AccountRepository,
	transactionRepository TransactionRepository,
) *AccountService {
	return &AccountService{
		accountRepository:     accountRepository,
		transactionRepository: transactionRepository,
	}
}

func (s *AccountService) GetAccount(ctx context.Context, userID int64) (models.Account, error) {
	return s.accountRepository.Get(ctx, userID)
}

func (s *AccountService) Withdraw(ctx context.Context, withdrawal models.Withdrawal) error {
	err := s.transactionRepository.Withdraw(ctx, withdrawal)

	if errors.Is(err, repositories.ErrNotEnoughMoney) {
		return models.ErrNotEnoughMoney
	}

	return err
}
