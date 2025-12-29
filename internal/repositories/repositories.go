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
	CreateTx(ctx context.Context, tx *sqlx.Tx, userID int64) error
	UpdateTx(ctx context.Context, tx *sqlx.Tx, userID int64, accrual, withdrawal int64) error
}

type UserRepository interface {
	CreateTx(ctx context.Context, tx *sqlx.Tx, user models.User) (int64, error)
}

type TransactionRepository struct {
	db          *sqlx.DB
	orderRepo   OrderRepository
	accountRepo AccountRepository
	userRepo    UserRepository
}

func NewTransactionRepository(
	db *sqlx.DB,
	orderRepo OrderRepository,
	accountRepo AccountRepository,
	userRepo UserRepository,
) *TransactionRepository {
	return &TransactionRepository{
		db:          db,
		orderRepo:   orderRepo,
		accountRepo: accountRepo,
		userRepo:    userRepo,
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

func (r *TransactionRepository) CreateUser(
	ctx context.Context,
	user models.User,
) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	userID, err := r.userRepo.CreateTx(ctx, tx, user)
	if err != nil {
		return -1, err
	}

	err = r.accountRepo.CreateTx(ctx, tx, userID)
	if err != nil {
		return -1, err
	}

	return userID, tx.Commit()
}
