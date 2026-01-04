package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// Config конфигурация для retry
type Config struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// Decorator декоратор для добавления retry логики
type Decorator struct {
	config Config
}

// NewRetryDecorator создает новый декоратор retry
func NewRetryDecorator(config Config) *Decorator {
	return &Decorator{config: config}
}

// Execute выполняет функцию с retry логикой
func (d *Decorator) Execute(fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= d.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := d.calculateDelay(attempt)
			time.Sleep(delay)
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Проверяем, является ли ошибка повторяемой
		if !d.isRetryableError(err) {
			return err
		}
	}

	return lastErr
}

// ExecuteWithTx выполняет функцию с retry логикой в транзакции
func (d *Decorator) ExecuteWithTx(ctx context.Context, fn func() error) error {
	return d.Execute(fn)
}

// isRetryableError проверяет, является ли ошибка повторяемой
func (d *Decorator) isRetryableError(err error) bool {
	// Транзиентные ошибки БД
	if err == sql.ErrConnDone || err == sql.ErrTxDone {
		return true
	}

	// PostgreSQL ошибки
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.SerializationFailure, // 40001
			pgerrcode.DeadlockDetected,                              // 40P01
			pgerrcode.AdminShutdown,                                 // 57P01
			pgerrcode.CrashShutdown,                                 // 57P02
			pgerrcode.CannotConnectNow,                              // 57P03
			pgerrcode.ConnectionException,                           // 08000
			pgerrcode.ConnectionFailure,                             // 08006
			pgerrcode.SQLClientUnableToEstablishSQLConnection,       // 08001
			pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection, // 08004
			pgerrcode.IdleSessionTimeout:                            // 25001 (пример)
			return true
		}
	}

	return false
}

// calculateDelay вычисляет задержку для retry
func (d *Decorator) calculateDelay(attempt int) time.Duration {
	delay := d.config.BaseDelay * time.Duration(1<<uint(attempt))
	if delay > d.config.MaxDelay {
		return d.config.MaxDelay
	}
	return delay
}
