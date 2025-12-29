package getorderhandler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/hydra13/gophermart/internal/models"
	authContext "github.com/hydra13/gophermart/internal/services/auth_context"
	"github.com/hydra13/gophermart/internal/utils"
)

type OrderService interface {
	GetOrdersByUser(ctx context.Context, userID int64) ([]models.Order, error)
}

type ResponseRecord struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"upload_at"`
}

type Handler struct {
	log zerolog.Logger
	o   OrderService
}

func NewHandler(o OrderService, log zerolog.Logger) *Handler {
	return &Handler{
		o:   o,
		log: log,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := authContext.GetUserIDFromContext(r.Context())

	orders, err := h.o.GetOrdersByUser(r.Context(), userID)
	if err != nil {
		h.log.Error().
			Err(err).
			Msg("error while getting orders")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(h.toResponse(orders))
	if err != nil {
		h.log.Error().
			Err(err).
			Msg("error encode response")
	}
}

func (h *Handler) toResponse(orders []models.Order) []ResponseRecord {
	result := make([]ResponseRecord, 0, len(orders))

	for _, order := range orders {
		result = append(result, ResponseRecord{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    utils.ToRub(order.Accrual),
			UploadedAt: order.UploadedAt,
		})
	}

	return result
}
