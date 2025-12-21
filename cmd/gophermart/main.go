package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	glog "go.finelli.dev/gooseloggers/zerolog"

	"github.com/hydra13/gophermart/internal/config"
	balanceHandler "github.com/hydra13/gophermart/internal/handlers/balance"
	getOrdersHandler "github.com/hydra13/gophermart/internal/handlers/get_orders"
	loadOrdersHandler "github.com/hydra13/gophermart/internal/handlers/load_orders"
	loginHandler "github.com/hydra13/gophermart/internal/handlers/login"
	registerHandler "github.com/hydra13/gophermart/internal/handlers/register"
	withdrawHandler "github.com/hydra13/gophermart/internal/handlers/withdraw"
	withdrawalsHandler "github.com/hydra13/gophermart/internal/handlers/withdrawals"
)

const dbDriver = "pgx"

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	conf := config.NewConfig()
	conf.Parse()

	dbInstance, err := sqlx.Connect(dbDriver, conf.DatabaseURI)
	if err != nil || dbInstance == nil {
		log.Fatal().
			Bool("dbInstanceIsNull", dbInstance == nil).
			Err(err).
			Msg("❌ Failed to connect to database")
	}
	defer dbInstance.Close()

	err = runMigrations(dbInstance, &log)
	if err != nil {
		log.Fatal().Err(err).Msg("❌ Failed to run migrations")
	}

	// Handlers
	getBalanceHandler := balanceHandler.NewHandler(log)
	getOrdersByUserHandler := getOrdersHandler.NewHandler(log)
	loadOrdersByUserHandler := loadOrdersHandler.NewHandler(log)
	loginHandler := loginHandler.NewHandler(log)
	registerHandler := registerHandler.NewHandler(log)
	withdrawHandler := withdrawHandler.NewHandler(log)
	withdrawalsHandler := withdrawalsHandler.NewHandler(log)

	// Routing
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

	srv := &http.Server{
		Addr:    conf.RunAddress,
		Handler: r,
	}

	go func() {
		log.Debug().Msg("🚀 Starting server at " + conf.RunAddress)
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("🟥 Server failed to start")
			} else {
				log.Info().Msg("🟨 Server closed")
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("⏳ Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("⚠️ Server forced to shutdown:")
	}
	log.Info().Msg("⬜ Server exited")
}

func runMigrations(conn *sqlx.DB, log *zerolog.Logger) error {
	goose.SetDialect(dbDriver)

	l := glog.GooseZerologLogger(log)
	goose.SetLogger(l)

	migrationsDir := "./migrations"
	if err := goose.Up(conn.DB, migrationsDir); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("✅ Migrations applied successfully")
	return nil
}
