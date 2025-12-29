package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/hydra13/gophermart/internal/models"
)

var (
	ErrOrderExists   = errors.New("order already exists")
	ErrOrderNotFound = errors.New("order not found")
	ErrConflict      = errors.New("conflict")
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, number string, userID int64) error {
	query := `
		INSERT INTO orders (number, user_id) VALUES ($1, $2)
		ON CONFLICT (number) DO NOTHING
	`
	result, err := r.db.ExecContext(ctx, query, number, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrOrderExists
	}
	return nil
}

func (r *OrderRepository) CheckAndCreate(ctx context.Context, number string, userID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userIDFromDB int64

	query := `SELECT user_id FROM orders WHERE number = $1 LIMIT 1`
	err = tx.GetContext(ctx, &userIDFromDB, query, number)
	if err != sql.ErrNoRows && err != nil {
		return err
	}

	if err != sql.ErrNoRows {
		if userID != userIDFromDB {
			return ErrConflict
		}

		return ErrOrderExists
	}

	query = `
		INSERT INTO orders (number, user_id) VALUES ($1, $2)
		ON CONFLICT (number) DO NOTHING
	`
	result, err := tx.ExecContext(ctx, query, number, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrOrderExists
	}

	return tx.Commit()
}

func (r *OrderRepository) GetByUserID(ctx context.Context, userID int64) ([]models.Order, error) {
	var orders []models.Order
	query := `
		SELECT number, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`
	err := r.db.SelectContext(ctx, &orders, query, userID)
	return orders, err
}

func (r *OrderRepository) GetByNumber(ctx context.Context, number string) (int64, error) {
	var userID int64
	query := `SELECT user_id FROM orders WHERE number = $1`
	err := r.db.GetContext(ctx, &userID, query, number)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrOrderNotFound
		}
		return 0, err
	}
	return userID, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, number string, status string, accrual int64) error {
	query := `
		UPDATE orders
		SET status = $1, accrual = $2
		WHERE number = $3
	`
	result, err := r.db.ExecContext(ctx, query, status, accrual, number)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrOrderNotFound
	}
	return nil
}

func (r *OrderRepository) GetProcessingOrders(ctx context.Context) ([]string, error) {
	var numbers []string
	query := `SELECT number FROM orders WHERE status = 'PROCESSING'`
	err := r.db.SelectContext(ctx, &numbers, query)
	return numbers, err
}
