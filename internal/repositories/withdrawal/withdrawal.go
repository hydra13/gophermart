package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/hydra13/gophermart/internal/models"
	repositories "github.com/hydra13/gophermart/internal/repositories"
)

type WithdrawalRepository struct {
	db    *sqlx.DB
	retry *repositories.Decorator
}

func NewWithdrawalRepository(db *sqlx.DB) *WithdrawalRepository {
	return &WithdrawalRepository{
		db: db,
		retry: repositories.NewRetryDecorator(repositories.Config{
			MaxRetries: repositories.RetryMaxRetries,
			BaseDelay:  repositories.RetryBaseDelay,
			MaxDelay:   repositories.RetryMaxDelay,
		}),
	}
}

func (r *WithdrawalRepository) AddTx(ctx context.Context, tx *sqlx.Tx, withdrawal models.Withdrawal) error {
	return r.retry.Execute(func() error {
		query := `
			INSERT INTO withdrawals (user_id, order_number, sum)
			VALUES ($1, $2, $3)
		`
		_, err := tx.ExecContext(ctx, query, withdrawal.UserID, withdrawal.OrderNumber, withdrawal.Sum)
		return err
	})
}

func (r *WithdrawalRepository) Create(ctx context.Context, userID int64, orderNumber string, sum int64) error {
	return r.retry.Execute(func() error {
		query := `
			INSERT INTO withdrawals (user_id, order_number, sum)
			VALUES ($1, $2, $3)
		`
		_, err := r.db.ExecContext(ctx, query, userID, orderNumber, sum)
		return err
	})
}

func (r *WithdrawalRepository) GetByUserID(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	var withdrawals []models.Withdrawal
	err := r.retry.Execute(func() error {
		query := `
			SELECT order_number, sum, processed_at
			FROM withdrawals
			WHERE user_id = $1
			ORDER BY processed_at DESC
		`
		return r.db.SelectContext(ctx, &withdrawals, query, userID)
	})
	return withdrawals, err
}

func (r *WithdrawalRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	return r.retry.Execute(func() error {
		query := `UPDATE withdrawals SET status = $1 WHERE id = $2`
		_, err := r.db.ExecContext(ctx, query, status, id)
		return err
	})
}
