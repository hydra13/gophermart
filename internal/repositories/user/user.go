package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	"github.com/hydra13/gophermart/internal/models"
	repositories "github.com/hydra13/gophermart/internal/repositories"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, login, passwordHash string) (int64, error) {
	var userID int64
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING user_id`
	err := r.db.QueryRowContext(ctx, query, login, passwordHash).Scan(&userID)
	if err != nil {
		// Проверка на дубликат логина (уникальный индекс)
		if isUniqueViolation(err) {
			return 0, repositories.ErrLoginExists
		}
		return 0, err
	}
	return userID, nil
}

func (r *UserRepository) CreateTx(ctx context.Context, tx *sqlx.Tx, user models.User) (int64, error) {
	var userID int64
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING user_id`
	err := tx.QueryRowContext(ctx, query, user.Login, user.PasswordHash).Scan(&userID)
	if err != nil {
		// Проверка на дубликат логина (уникальный индекс)
		if isUniqueViolation(err) {
			return 0, repositories.ErrLoginExists
		}
		return 0, err
	}
	return userID, nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	query := `SELECT user_id, login, password_hash, created_at FROM users WHERE login = $1`
	err := r.db.GetContext(ctx, &user, query, login)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repositories.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgerrcode.UniqueViolation
	}
	return false
}
