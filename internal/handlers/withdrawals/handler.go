package withdrawals

import (
	"net/http"

	"github.com/rs/zerolog"
)

type Handler struct {
	log zerolog.Logger
}

func NewHandler(log zerolog.Logger) *Handler {
	return &Handler{
		log: log,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
}
