//go:generate minimock -i .OrderService -o mocks -s _mock.go -g
package addorderhandler

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/hydra13/gophermart/internal/models"
	authContext "github.com/hydra13/gophermart/internal/services/auth_context"
	"github.com/hydra13/gophermart/internal/validators"
)

type OrderService interface {
	AddOrder(ctx context.Context, orderNumber string, userID int64) error
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
	if r.Header.Get("Content-Type") != "text/plain" {
		h.log.Error().
			Str("content-type", r.Header.Get("Content-Type")).
			Msg("add_order: invalid content type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Error().
			Err(err).
			Msg("add_order: can't read body")

		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNum := string(body)

	if orderNum == "" {
		h.log.Error().
			Msg("add_order: empty order number")

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.log.Debug().
		Str("order", orderNum).
		Msg("add_order: new order")

	if !validators.IsValidLuhn(orderNum) {
		h.log.Error().
			Str("order", orderNum).
			Msg("add_order: invalid order number")

		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userID := authContext.GetUserIDFromContext(r.Context())

	err = h.o.AddOrder(r.Context(), orderNum, userID)

	if errors.Is(err, models.ErrOrderAlreadyExists) {
		h.log.Info().
			Str("order", orderNum).
			Msg("add_order: order already exists")

		w.WriteHeader(http.StatusOK)
		return
	}

	if errors.Is(err, models.ErrConflict) {
		h.log.Error().
			Err(err).
			Msg("add_order: order already added by another user")

		w.WriteHeader(http.StatusConflict)
		return
	}

	if err != nil {

		h.log.Error().
			Err(err).
			Msg("add_order: can't add order")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
