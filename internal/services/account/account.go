package account

import (
	"context"

	"github.com/hydra13/gophermart/internal/models"
)

type AccountRepository interface {
	Get(ctx context.Context, userID int64) (models.Account, error)
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
	// TODO: не прокидывать наружу ошибки репозитория
	return s.accountRepository.Get(ctx, userID)
}
