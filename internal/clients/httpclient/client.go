package httpclient

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const (
	maxRetries       = 3
	baseDelay        = time.Millisecond * 100
	maxDelay         = time.Second
	defaultRateLimit = rate.Limit(10)
	defaultBurst     = 10
)

type Client struct {
	client  *http.Client
	limiter *rate.Limiter
}

func New() *Client {
	return &Client{
		client:  &http.Client{},
		limiter: rate.NewLimiter(defaultRateLimit, defaultBurst),
	}
}

func NewWithParams(rateLimit rate.Limit, burst int) *Client {
	return &Client{
		client:  &http.Client{},
		limiter: rate.NewLimiter(rateLimit, burst),
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		ctx := req.Context()
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http request error: %w", err)

			// Если это последняя попытка, возвращаем ошибку
			if attempt == maxRetries {
				return nil, lastErr
			}

			delay := c.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}

			continue
		}

		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("HTTP error: %d", resp.StatusCode)

			if attempt == maxRetries {
				resp.Body.Close()

				return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
			}

			delay := c.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}

			resp.Body.Close()
			continue
		}

		return resp, nil
	}

	return nil, lastErr
}

// Рассчет задержки с экспоненциальным увеличением
func (c *Client) calculateDelay(attempt int) time.Duration {
	delay := baseDelay * time.Duration(1<<uint(attempt))
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

func (c *Client) SetRateLimit(rateLimit rate.Limit, burst int) {
	c.limiter.SetLimit(rateLimit)
	c.limiter.SetBurst(burst)
}
