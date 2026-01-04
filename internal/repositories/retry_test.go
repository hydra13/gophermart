package repositories

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestDecorator_Execute(t *testing.T) {
	tests := []struct {
		name          string
		maxRetries    int
		errors        []error
		expectedCalls int
		expectError   bool
		expectedError string
	}{
		{
			name:          "success on first attempt",
			maxRetries:    3,
			errors:        []error{nil},
			expectedCalls: 1,
			expectError:   false,
		},
		{
			name:          "success after retry",
			maxRetries:    3,
			errors:        []error{&pgconn.PgError{Code: pgerrcode.ConnectionException}, nil},
			expectedCalls: 2,
			expectError:   false,
		},
		{
			name:          "max retries exceeded",
			maxRetries:    2,
			errors:        []error{&pgconn.PgError{Code: pgerrcode.ConnectionException}, &pgconn.PgError{Code: pgerrcode.ConnectionException}, &pgconn.PgError{Code: pgerrcode.ConnectionException}},
			expectedCalls: 3,
			expectError:   true,
		},
		{
			name:          "non-retryable error stops retry",
			maxRetries:    3,
			errors:        []error{errors.New("non-retryable error")},
			expectedCalls: 1,
			expectError:   true,
			expectedError: "non-retryable error",
		},
		{
			name:          "mixed retryable and non-retryable errors",
			maxRetries:    3,
			errors:        []error{&pgconn.PgError{Code: pgerrcode.ConnectionException}, errors.New("non-retryable error")},
			expectedCalls: 2,
			expectError:   true,
			expectedError: "non-retryable error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decorator := NewRetryDecorator(Config{
				MaxRetries: tt.maxRetries,
				BaseDelay:  time.Microsecond, // Very short delay for tests
				MaxDelay:   time.Microsecond * 10,
			})

			callCount := 0
			err := decorator.Execute(func() error {
				callCount++
				if callCount <= len(tt.errors) {
					return tt.errors[callCount-1]
				}
				return nil
			})

			assert.Equal(t, tt.expectedCalls, callCount)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDecorator_isRetryableError(t *testing.T) {
	decorator := NewRetryDecorator(Config{})

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "sql.ErrConnDone",
			err:      sql.ErrConnDone,
			expected: true,
		},
		{
			name:     "sql.ErrTxDone",
			err:      sql.ErrTxDone,
			expected: true,
		},
		{
			name: "serialization failure",
			err: &pgconn.PgError{
				Code: pgerrcode.SerializationFailure,
			},
			expected: true,
		},
		{
			name: "deadlock detected",
			err: &pgconn.PgError{
				Code: pgerrcode.DeadlockDetected,
			},
			expected: true,
		},
		{
			name: "connection exception",
			err: &pgconn.PgError{
				Code: pgerrcode.ConnectionException,
			},
			expected: true,
		},
		{
			name: "non-retryable error",
			err: &pgconn.PgError{
				Code: "23505", // unique_violation
			},
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decorator.isRetryableError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecorator_calculateDelay(t *testing.T) {
	decorator := NewRetryDecorator(Config{
		BaseDelay: time.Millisecond,
		MaxDelay:  time.Millisecond * 100,
	})

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{attempt: 0, expected: time.Millisecond},
		{attempt: 1, expected: time.Millisecond * 2},
		{attempt: 2, expected: time.Millisecond * 4},
		{attempt: 3, expected: time.Millisecond * 8},
		{attempt: 4, expected: time.Millisecond * 16},
		{attempt: 7, expected: time.Millisecond * 100}, // Max delay reached
	}

	for _, tt := range tests {
		result := decorator.calculateDelay(tt.attempt)
		assert.Equal(t, tt.expected, result)
	}
}
