package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

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

var (
	ErrAccualServerResponseStatus      = errors.New("accrual server return unexpected status")
	ErrAccualServerResponseContentType = errors.New("accrual server return unexpected content type")
	ErrAccualServerTooManyRequests     = errors.New("accrual server return too many requests")
	ErrAccualServerInternalError       = errors.New("accrual server return internal server error")
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	accrualSystemAddress string
	httpClient           HTTPClient
}

func New(accrualSystemAddress string, httpClient HTTPClient) Client {
	return Client{
		accrualSystemAddress: accrualSystemAddress,
		httpClient:           httpClient,
	}
}

func (c Client) GetOrderStatus(ctx context.Context, orderNumber string) (models.AccrualResponse, error) {
	// Создаем запрос
	requestURL := c.generateRequestURL(orderNumber)
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return models.AccrualResponse{}, err
	}

	// Выполняем запрос через HTTP клиент с ретраями и rate limiting
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return models.AccrualResponse{}, err
	}

	defer resp.Body.Close()

	// Обрабатываем ответ
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return c.parseSuccessfulResponse(resp, orderNumber)
	}

	// 204 No Content
	if resp.StatusCode == http.StatusNoContent {
		return models.AccrualResponse{}, models.ErrOrderNotFound
	}

	// Для других ошибок
	return models.AccrualResponse{}, fmt.Errorf("HTTP error: %d", resp.StatusCode)
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
