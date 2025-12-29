package account

import (
	"context"

	"github.com/hydra13/gophermart/internal/models"
	"github.com/hydra13/gophermart/internal/repositories"
)

type AccountRepository interface {
	Get(ctx context.Context, userID int64) (models.Account, error)
	Withdraw(ctx context.Context, userID int64, amount int64) error
}

type AccountService struct {
	accountRepository AccountRepository
}

func NewAccountService(accountRepository AccountRepository) *AccountService {
	return &AccountService{
		accountRepository: accountRepository,
	}
}

func (s *AccountService) GetAccount(ctx context.Context, userID int64) (models.Account, error) {
	// TODO: надо мапить на ошибки сервисного слоя а не прокидывать ошибки репозитория
	return s.accountRepository.Get(ctx, userID)
}

func (s *AccountService) Withdraw(ctx context.Context, userID int64, amount int64) error {
	// TODO: надо мапить на ошибки сервисного слоя а не прокидывать ошибки репозитория
	err := s.accountRepository.Withdraw(ctx, userID, amount)

	if err == repositories.ErrNotEnoughMoney {
		return models.ErrNotEnoughMoney
	}

	return err
}
