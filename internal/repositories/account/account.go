package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/hydra13/gophermart/internal/models"
)

var ErrAccountNotFound = errors.New("account not found")

type AccountRepository struct {
	db *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Get(ctx context.Context, userID int64) (*models.Account, error) {
	var acc models.Account
	query := `SELECT user_id, current, withdrawn FROM accounts WHERE user_id = $1`
	err := r.db.GetContext(ctx, &acc, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return &acc, nil
}

func (r *AccountRepository) Update(ctx context.Context, userID int64, accrual, withdrawal int64) error {
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
		return ErrAccountNotFound
	}
	return nil
}
