package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type Config struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

type Decorator struct {
	config Config
}

func NewRetryDecorator(config Config) *Decorator {
	return &Decorator{config: config}
}

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

		if !d.isRetryableError(err) {
			return err
		}
	}

	return lastErr
}

func (d *Decorator) ExecuteWithTx(ctx context.Context, fn func() error) error {
	return d.Execute(fn)
}

func (d *Decorator) isRetryableError(err error) bool {
	// Транзиентные ошибки БД
	if err == sql.ErrConnDone || err == sql.ErrTxDone {
		return true
	}

	// PostgreSQL ошибки
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.SerializationFailure,
			pgerrcode.DeadlockDetected,
			pgerrcode.AdminShutdown,
			pgerrcode.CrashShutdown,
			pgerrcode.CannotConnectNow,
			pgerrcode.ConnectionException,
			pgerrcode.ConnectionFailure,
			pgerrcode.SQLClientUnableToEstablishSQLConnection,
			pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			pgerrcode.IdleSessionTimeout:
			return true
		}
	}

	return false
}

func (d *Decorator) calculateDelay(attempt int) time.Duration {
	delay := d.config.BaseDelay * time.Duration(1<<uint(attempt))
	if delay > d.config.MaxDelay {
		return d.config.MaxDelay
	}
	return delay
}
