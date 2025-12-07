package main

import (
	"context"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/hydra13/gophermart/internal/config"
	balanceHandler "github.com/hydra13/gophermart/internal/handlers/balance"
	getOrdersHandler "github.com/hydra13/gophermart/internal/handlers/get_orders"
	loadOrdersHandler "github.com/hydra13/gophermart/internal/handlers/load_orders"
	loginHandler "github.com/hydra13/gophermart/internal/handlers/login"
	registerHandler "github.com/hydra13/gophermart/internal/handlers/register"
	withdrawHandler "github.com/hydra13/gophermart/internal/handlers/withdraw"
	withdrawalsHandler "github.com/hydra13/gophermart/internal/handlers/withdrawals"
)

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf := config.NewConfig()
	conf.Parse()

	getBalanceHandler := balanceHandler.NewHandler(log)
	getOrdersByUserHandler := getOrdersHandler.NewHandler(log)
	loadOrdersByUserHandler := loadOrdersHandler.NewHandler(log)
	loginHandler := loginHandler.NewHandler(log)
	registerHandler := registerHandler.NewHandler(log)
	withdrawHandler := withdrawHandler.NewHandler(log)
	withdrawalsHandler := withdrawalsHandler.NewHandler(log)

	r := chi.NewRouter()

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", registerHandler.Handle)
		r.Post("/login", loginHandler.Handle)
		r.Post("/orders", loadOrdersByUserHandler.Handle)
		r.Get("/orders", getOrdersByUserHandler.Handle)
		r.Route("/balance", func(r chi.Router) {
			r.Get("/", getBalanceHandler.Handle)
			r.Post("/withdraw", withdrawHandler.Handle)
		})
		r.Get("/withdrawals", withdrawalsHandler.Handle)
	})
}
