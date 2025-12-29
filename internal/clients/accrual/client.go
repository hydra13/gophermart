package accrual

import (
	"context"
	"encoding/json"
	"errors"
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
	requestURL := c.generateRequestURL(orderNumber)

	resp, err := http.Get(requestURL)
	if err != nil {
		return models.AccrualResponse{}, err
	}

	defer resp.Body.Close()

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
