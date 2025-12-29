package repositories

import (
	"context"

	"github.com/hydra13/gophermart/internal/models"
	"github.com/jmoiron/sqlx"
)

type OrderRepository interface {
	UpdateTx(ctx context.Context, tx *sqlx.Tx, number string, status string, accrual int64) error
}

type AccountRepository interface {
	UpdateTx(ctx context.Context, tx *sqlx.Tx, userID int64, accrual, withdrawal int64) error
}

type TransactionRepository struct {
	db          *sqlx.DB
	orderRepo   OrderRepository
	accountRepo AccountRepository
}

func NewTransactionRepository(
	db *sqlx.DB,
	orderRepo OrderRepository,
	accountRepo AccountRepository,
) *TransactionRepository {
	return &TransactionRepository{
		db:          db,
		orderRepo:   orderRepo,
		accountRepo: accountRepo,
	}
}

func (r *TransactionRepository) UpdateOrderAndAccount(
	ctx context.Context,
	order models.Order,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = r.orderRepo.UpdateTx(ctx, tx, order.Number, order.Status, order.Accrual)
	if err != nil {
		return err
	}

	err = r.accountRepo.UpdateTx(ctx, tx, order.UserID, order.Accrual, 0)
	if err != nil {
		return err
	}

	return tx.Commit()
}
