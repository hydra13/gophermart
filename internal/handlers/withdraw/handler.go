//go:generate minimock -i .AccountService -o mocks -s _mock.go -g
package withdrawhandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/hydra13/gophermart/internal/models"
	authContext "github.com/hydra13/gophermart/internal/services/auth_context"
	"github.com/hydra13/gophermart/internal/utils"
	"github.com/hydra13/gophermart/internal/validators"
)

type JSONRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type AccountService interface {
	Withdraw(ctx context.Context, withdrawal models.Withdrawal) error
}

type Handler struct {
	accountService AccountService
	log            zerolog.Logger
}

func NewHandler(accountService AccountService, log zerolog.Logger) *Handler {
	return &Handler{
		accountService: accountService,
		log:            log,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := authContext.GetUserIDFromContext(r.Context())

	if r.Header.Get("Content-Type") != "application/json" {
		h.log.Debug().
			Str("content-type", r.Header.Get("Content-Type")).
			Msg("withdraw: error by content-type")

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req JSONRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.log.Debug().
			Err(err).
			Msg("withdraw: error read request body")

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !validators.IsValidLuhn(req.Order) {
		h.log.Debug().
			Str("order_number", req.Order).
			Msg("withdraw: validation error by order number")

		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	sum := utils.FromRub(req.Sum)

	if !validators.IsValidWithdrawAmount(sum) {
		h.log.Debug().
			Str("order_number", req.Order).
			Float64("sum", req.Sum).
			Msg("withdraw: validation error by withdraw amount")

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.accountService.Withdraw(r.Context(), models.Withdrawal{
		OrderNumber: req.Order,
		Sum:         sum,
		UserID:      userID,
	})
	if err == models.ErrNotEnoughMoney {
		h.log.Debug().
			Err(err).
			Msg("withdraw: error not enough money")
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}
	if err != nil {
		h.log.Debug().
			Err(err).
			Msg("withdraw: error withdraw")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
