//go:generate minimock -i .AccountService -o mocks -s _mock.go -g
package balancehandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/hydra13/gophermart/internal/models"
	authContext "github.com/hydra13/gophermart/internal/services/auth_context"
	"github.com/hydra13/gophermart/internal/utils"
)

type JSONResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type AccountService interface {
	GetAccount(ctx context.Context, userID int64) (models.Account, error)
}

type Handler struct {
	accountService AccountService
	log            zerolog.Logger
}

func NewHandler(
	accountService AccountService,
	log zerolog.Logger,
) *Handler {
	return &Handler{
		accountService: accountService,
		log:            log,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := authContext.GetUserIDFromContext(r.Context())

	account, err := h.accountService.GetAccount(r.Context(), userID)
	if err != nil {
		h.log.Error().
			Err(err).
			Msg("failed to get account")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(h.toResponse(account))
}

func (h *Handler) toResponse(account models.Account) JSONResponse {
	return JSONResponse{
		Current:   utils.ToRub(account.Current),
		Withdrawn: utils.ToRub(account.Withdrawn),
	}
}
