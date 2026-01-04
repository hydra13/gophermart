package accrual

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hydra13/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestClient_GetOrderStatus(t *testing.T) {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		setup        func() context.Context
		orderNumber  string
		wantErr      bool
		expectedResp *models.AccrualResponse
		errorIs      error
		errorEqual   error
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"order": "12345", "status": "PROCESSED", "accrual": 100.5}`))
			},
			setup:       func() context.Context { return context.Background() },
			orderNumber: "12345",
			wantErr:     false,
			expectedResp: &models.AccrualResponse{
				Order:   "12345",
				Status:  models.OrderStatusProcessed,
				Accrual: 10050,
			},
		},
		{
			name: "status no content",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			setup:       func() context.Context { return context.Background() },
			orderNumber: "12345",
			wantErr:     true,
			errorIs:     models.ErrOrderNotFound,
		},
		{
			name: "invalid content type",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid response"))
			},
			setup:       func() context.Context { return context.Background() },
			orderNumber: "12345",
			wantErr:     true,
			errorIs:     ErrAccualServerResponseContentType,
		},
		{
			name: "too many requests error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
			},
			setup:       func() context.Context { return context.Background() },
			orderNumber: "12345",
			wantErr:     true,
			errorIs:     ErrAccualServerTooManyRequests,
		},
		{
			name: "server error 500",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			setup:       func() context.Context { return context.Background() },
			orderNumber: "12345",
			wantErr:     true,
		},
		{
			name: "context cancelled",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"order": "12345", "status": "PROCESSED", "accrual": 100.5}`))
			},
			setup: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			orderNumber: "12345",
			wantErr:     true,
			errorEqual:  context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := New(server.URL)
			ctx := tt.setup()

			resp, err := client.GetOrderStatus(ctx, tt.orderNumber)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorIs != nil {
					assert.True(t, errors.Is(err, tt.errorIs))
				}
				if tt.errorEqual != nil {
					assert.Equal(t, tt.errorEqual, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp.Order, resp.Order)
				assert.Equal(t, tt.expectedResp.Status, resp.Status)
				assert.Equal(t, tt.expectedResp.Accrual, resp.Accrual)
			}
		})
	}
}

func TestClient_GetOrderStatus_RetryOn500(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"order": "12345", "status": "PROCESSED", "accrual": 100.5}`))
	}))
	defer server.Close()

	client := New(server.URL)
	resp, err := client.GetOrderStatus(context.Background(), "12345")

	assert.NoError(t, err)
	assert.Equal(t, "12345", resp.Order)
	assert.Equal(t, models.OrderStatusProcessed, resp.Status)
	assert.Equal(t, int64(10050), resp.Accrual)
	assert.Greater(t, attempts, 1)
}

func TestClient_GetOrderStatus_TooManyRequests(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"order": "12345", "status": "PROCESSED", "accrual": 100.5}`))
	}))
	defer server.Close()

	client := New(server.URL)
	resp, err := client.GetOrderStatus(context.Background(), "12345")

	assert.NoError(t, err)
	assert.Equal(t, "12345", resp.Order)
	assert.Equal(t, models.OrderStatusProcessed, resp.Status)
	assert.Equal(t, int64(10050), resp.Accrual)
	assert.Greater(t, attempts, 1)
}
