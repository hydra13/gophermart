package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/hydra13/gophermart/internal/models"
	repositories "github.com/hydra13/gophermart/internal/repositories"
)

type OrderRepository struct {
	db    *sqlx.DB
	retry *repositories.Decorator
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
		retry: repositories.NewRetryDecorator(repositories.Config{
			MaxRetries: repositories.RetryMaxRetries,
			BaseDelay:  repositories.RetryBaseDelay,
			MaxDelay:   repositories.RetryMaxDelay,
		}),
	}
}

func (r *OrderRepository) Create(ctx context.Context, number string, userID int64) error {
	return r.retry.Execute(func() error {
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
			return repositories.ErrOrderExists
		}
		return nil
	})
}

func (r *OrderRepository) CheckAndCreate(ctx context.Context, number string, userID int64) error {
	return r.retry.Execute(func() error {
		tx, err := r.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		var userIDFromDB int64

		query := `SELECT user_id FROM orders WHERE number = $1 LIMIT 1`
		err = tx.GetContext(ctx, &userIDFromDB, query, number)
		if !errors.Is(err, sql.ErrNoRows) && err != nil {
			return err
		}

		if !errors.Is(err, sql.ErrNoRows) {
			if userID != userIDFromDB {
				return repositories.ErrConflict
			}

			return repositories.ErrOrderExists
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
			return repositories.ErrOrderExists
		}

		return tx.Commit()
	})
}

func (r *OrderRepository) GetByUserID(ctx context.Context, userID int64) ([]models.Order, error) {
	var orders []models.Order
	err := r.retry.Execute(func() error {
		query := `
			SELECT number, status, accrual, uploaded_at
			FROM orders
			WHERE user_id = $1
			ORDER BY uploaded_at DESC
		`
		err := r.db.SelectContext(ctx, &orders, query, userID)
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	})
	return orders, err
}

func (r *OrderRepository) GetByNumber(ctx context.Context, number string) (int64, error) {
	var userID int64
	err := r.retry.Execute(func() error {
		query := `SELECT user_id FROM orders WHERE number = $1`
		err := r.db.GetContext(ctx, &userID, query, number)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return repositories.ErrOrderNotFound
			}
			return err
		}
		return nil
	})
	return userID, err
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, number string, status string, accrual int64) error {
	return r.retry.Execute(func() error {
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
			return repositories.ErrOrderNotFound
		}
		return nil
	})
}

func (r *OrderRepository) UpdateTx(
	ctx context.Context,
	tx *sqlx.Tx,
	number string,
	status string,
	accrual int64,
) error {
	return r.retry.Execute(func() error {
		query := `
			UPDATE orders
			SET status = $1, accrual = $2
			WHERE number = $3
		`
		result, err := tx.ExecContext(ctx, query, status, accrual, number)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return repositories.ErrOrderNotFound
		}
		return nil
	})
}

func (r *OrderRepository) GetOrdersForCheckStatus(ctx context.Context) ([]models.Order, error) {
	var orders []models.Order
	err := r.retry.Execute(func() error {
		query := `SELECT number, user_id FROM orders WHERE status = 'NEW' or status = 'PROCESSING'`
		return r.db.SelectContext(ctx, &orders, query)
	})
	return orders, err
}
