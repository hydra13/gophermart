package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/hydra13/gophermart/internal/models"
	"github.com/hydra13/gophermart/internal/utils"
)

type JSONResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

const (
	AccualStatusRegistered = "REGISTERED"
	AccualStatusProcessing = "PROCESSING"
	AccualStatusInvalid    = "INVALID"
	AccualStatusProcessed  = "PROCESSED"
)

const (
	maxRetries = 3
	baseDelay  = time.Millisecond * 100
	maxDelay   = time.Second
)

var (
	ErrAccualServerResponseStatus      = errors.New("accrual server return unexpected status")
	ErrAccualServerResponseContentType = errors.New("accrual server return unexpected content type")
	ErrAccualServerTooManyRequests     = errors.New("too many requests")
)

type Client struct {
	accrualSystemAddress string
}

func New(accrualSystemAddress string) Client {
	return Client{
		accrualSystemAddress: accrualSystemAddress,
	}
}

func (c Client) GetOrderStatus(ctx context.Context, orderNumber string) (models.AccrualResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := c.calculateDelay(attempt)

			select {
			case <-ctx.Done():
				return models.AccrualResponse{}, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, err := c.makeRequest(ctx, orderNumber)
		if err != nil {
			lastErr = err
			continue
		}

		defer resp.Body.Close()

		// Успешный ответ
		if resp.StatusCode >= 200 && resp.StatusCode < 300 && resp.StatusCode != http.StatusTooManyRequests {
			return c.parseSuccessfulResponse(resp, orderNumber)
		}

		// 429 Too Many Requests
		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = ErrAccualServerTooManyRequests
			// Если есть заголовок Retry-After, используем его
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil {
					select {
					case <-ctx.Done():
						return models.AccrualResponse{}, ctx.Err()
					case <-time.After(time.Duration(seconds) * time.Second):
					}
				}
			}
			continue
		}

		// 5xx - обработка ошибок сервера
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		// 204 No Content
		if resp.StatusCode == http.StatusNoContent {
			return models.AccrualResponse{}, models.ErrOrderNotFound
		}

		// Для других ошибок не повторяем
		lastErr = fmt.Errorf("HTTP error: %d", resp.StatusCode)
		break
	}

	return models.AccrualResponse{}, lastErr
}

func (c Client) calculateDelay(attempt int) time.Duration {
	delay := baseDelay * time.Duration(1<<uint(attempt))
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

func (c Client) parseSuccessfulResponse(resp *http.Response, orderNumber string) (models.AccrualResponse, error) {
	if resp.StatusCode == http.StatusNoContent {
		return models.AccrualResponse{}, models.ErrOrderNotFound
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		return models.AccrualResponse{}, ErrAccualServerResponseContentType
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.AccrualResponse{}, err
	}

	var respJSON JSONResponse
	err = json.Unmarshal(body, &respJSON)
	if err != nil {
		return models.AccrualResponse{}, err
	}

	status, err := c.mapStatus(respJSON.Status)
	if err != nil {
		return models.AccrualResponse{}, err
	}

	return models.AccrualResponse{
		Order:   orderNumber,
		Status:  status,
		Accrual: utils.FromRub(respJSON.Accrual),
	}, nil
}

func (c Client) makeRequest(ctx context.Context, orderNumber string) (*http.Response, error) {
	requestURL := c.generateRequestURL(orderNumber)
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}

func (c *Client) generateRequestURL(orderNumber string) string {
	return c.accrualSystemAddress + "/api/orders/" + orderNumber
}

func (c *Client) mapStatus(accualStatus string) (string, error) {
	switch accualStatus {
	case AccualStatusRegistered:
		return models.OrderStatusProcessing, nil
	case AccualStatusProcessing:
		return models.OrderStatusProcessing, nil
	case AccualStatusInvalid:
		return models.OrderStatusInvalid, nil
	case AccualStatusProcessed:
		return models.OrderStatusProcessed, nil
	default:
		return "", ErrAccualServerResponseStatus
	}
}
