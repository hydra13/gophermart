package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/hydra13/gophermart/internal/models"
)

type WithdrawalRepository struct {
	db *sqlx.DB
}

func NewWithdrawalRepository(db *sqlx.DB) *WithdrawalRepository {
	return &WithdrawalRepository{db: db}
}

func (r *WithdrawalRepository) Create(ctx context.Context, userID int64, orderNumber string, sum int64) error {
	query := `
		INSERT INTO withdrawals (user_id, order_number, sum)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, userID, orderNumber, sum)
	return err
}

func (r *WithdrawalRepository) GetByUserID(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	var withdrawals []models.Withdrawal
	query := `
		SELECT order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`
	err := r.db.SelectContext(ctx, &withdrawals, query, userID)
	return withdrawals, err
}

func (r *WithdrawalRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE withdrawals SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}
