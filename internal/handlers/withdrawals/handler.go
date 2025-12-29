//go:generate minimock -i .WithdrawalService -o mocks -s _mock.go -g
package withdrawals

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

type WithdrawalService interface {
	GetWithdrawalsByUser(ctx context.Context, userID int64) ([]models.Withdrawal, error)
}

type ResponseRecord struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type Handler struct {
	withdrawalService WithdrawalService
	log               zerolog.Logger
}

func NewHandler(
	withdrawalService WithdrawalService,
	log zerolog.Logger,
) *Handler {
	return &Handler{
		withdrawalService: withdrawalService,
		log:               log,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := authContext.GetUserIDFromContext(r.Context())

	withdrawals, err := h.withdrawalService.GetWithdrawalsByUser(r.Context(), userID)
	if err == models.ErrNoWithdrawals {
		h.log.Debug().
			Int64("user_id", userID).
			Msg("withdrawals: no withdrawals for user")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		h.log.Error().
			Err(err).
			Msg("withdrawals: error while getting withdrawals by user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(h.toResponse(withdrawals))
	if err != nil {
		h.log.Error().
			Err(err).
			Msg("withdrawals: error encode response")
	}
}

func (h *Handler) toResponse(withdrawals []models.Withdrawal) []ResponseRecord {
	records := make([]ResponseRecord, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		records = append(records, ResponseRecord{
			Order:       withdrawal.OrderNumber,
			Sum:         utils.ToRub(withdrawal.Sum),
			ProcessedAt: withdrawal.ProcessedAt,
		})
	}
	return records
}
