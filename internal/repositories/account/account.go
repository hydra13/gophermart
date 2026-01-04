package repositories

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/hydra13/gophermart/internal/models"
	repositories "github.com/hydra13/gophermart/internal/repositories"
)

type AccountRepository struct {
	db    *sqlx.DB
	retry *repositories.Decorator
}

func NewAccountRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{
		db: db,
		retry: repositories.NewRetryDecorator(repositories.Config{
			MaxRetries: repositories.RetryMaxRetries,
			BaseDelay:  repositories.RetryBaseDelay,
			MaxDelay:   repositories.RetryMaxDelay,
		}),
	}
}

func (r *AccountRepository) CreateTx(ctx context.Context, tx *sqlx.Tx, userID int64) error {
	return r.retry.Execute(func() error {
		query := `INSERT INTO accounts (user_id, current, withdrawn) VALUES ($1, 0, 0)`
		_, err := tx.ExecContext(ctx, query, userID)
		return err
	})
}

func (r *AccountRepository) Get(ctx context.Context, userID int64) (models.Account, error) {
	var acc models.Account
	err := r.retry.Execute(func() error {
		query := `SELECT user_id, current, withdrawn FROM accounts WHERE user_id = $1`
		err := r.db.GetContext(ctx, &acc, query, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				return repositories.ErrAccountNotFound
			}
			return err
		}
		return nil
	})
	return acc, err
}

func (r *AccountRepository) Update(ctx context.Context, userID int64, accrual, withdrawal int64) error {
	return r.retry.Execute(func() error {
		query := `
			UPDATE accounts
			SET current = current + $1, withdrawn = withdrawn + $2
			WHERE user_id = $3
		`
		result, err := r.db.ExecContext(ctx, query, accrual, withdrawal, userID)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return repositories.ErrAccountNotFound
		}
		return nil
	})
}

func (r *AccountRepository) UpdateTx(ctx context.Context, tx *sqlx.Tx, userID int64, accrual, withdrawal int64) error {
	return r.retry.Execute(func() error {
		query := `
			UPDATE accounts
			SET current = current + $1, withdrawn = withdrawn + $2
			WHERE user_id = $3
		`
		result, err := tx.ExecContext(ctx, query, accrual, withdrawal, userID)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return repositories.ErrAccountNotFound
		}
		return nil
	})
}

func (r *AccountRepository) Withdraw(ctx context.Context, userID int64, amount int64) error {
	return r.retry.Execute(func() error {
		queryUpdate := `
			UPDATE accounts
			SET current = current - $1, withdrawn = withdrawn + $1
			WHERE user_id = $2 AND current >= $1
		`
		result, err := r.db.ExecContext(ctx, queryUpdate, amount, userID)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return repositories.ErrNotEnoughMoney
		}
		return nil
	})
}

func (r *AccountRepository) WithdrawTx(ctx context.Context, tx *sqlx.Tx, userID int64, amount int64) error {
	return r.retry.Execute(func() error {
		queryUpdate := `
			UPDATE accounts
			SET current = current - $1, withdrawn = withdrawn + $1
			WHERE user_id = $2 AND current >= $1
		`
		result, err := tx.ExecContext(ctx, queryUpdate, amount, userID)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return repositories.ErrNotEnoughMoney
		}
		return nil
	})
}
