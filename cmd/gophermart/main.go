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

	accrualClient "github.com/hydra13/gophermart/internal/clients/accrual"
	"github.com/hydra13/gophermart/internal/config"
	addOrderHandler "github.com/hydra13/gophermart/internal/handlers/add_order"
	balanceHandler "github.com/hydra13/gophermart/internal/handlers/balance"
	getOrdersHandler "github.com/hydra13/gophermart/internal/handlers/get_orders"
	loginHandler "github.com/hydra13/gophermart/internal/handlers/login"
	registerHandler "github.com/hydra13/gophermart/internal/handlers/register"
	withdrawHandler "github.com/hydra13/gophermart/internal/handlers/withdraw"
	withdrawalsHandler "github.com/hydra13/gophermart/internal/handlers/withdrawals"
	authMiddleware "github.com/hydra13/gophermart/internal/middlewares/auth"
	"github.com/hydra13/gophermart/internal/middlewares/compresser"
	"github.com/hydra13/gophermart/internal/middlewares/logger"
	orderRepository "github.com/hydra13/gophermart/internal/repositories/order"
	userRepository "github.com/hydra13/gophermart/internal/repositories/user"
	authService "github.com/hydra13/gophermart/internal/services/auth"
	orderService "github.com/hydra13/gophermart/internal/services/order"
	userService "github.com/hydra13/gophermart/internal/services/user"
	"github.com/hydra13/gophermart/internal/services/worker"
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

	// Clients
	accrual := accrualClient.New(conf.AccrualSystemAddress)

	// Services
	auth := authService.New()
	userRepo := userRepository.NewUserRepository(dbInstance)
	orderRepo := orderRepository.NewOrderRepository(dbInstance)
	user := userService.NewUserService(userRepo)
	order := orderService.NewOrderService(orderRepo)
	w := worker.NewWorker(accrual, order, log)

	//Middlewares
	authMiddleware := authMiddleware.NewAuthMiddleware(auth, log)

	// Handlers
	getBalanceHandler := balanceHandler.NewHandler(log)
	getOrdersByUserHandler := getOrdersHandler.NewHandler(log)
	addOrderByUserHandler := addOrderHandler.NewHandler(order, log)
	loginHandler := loginHandler.NewHandler(user, auth, log)
	registerHandler := registerHandler.NewHandler(user, auth, log)
	withdrawHandler := withdrawHandler.NewHandler(log)
	withdrawalsHandler := withdrawalsHandler.NewHandler(log)

	// Routing
	r := chi.NewRouter()

	r.Use(compresser.CompresserMiddleware)
	r.Use(logger.NewLoggerMiddleware(log))

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", registerHandler.Handle)
		r.Post("/login", loginHandler.Handle)
		r.Route("/orders", func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/", addOrderByUserHandler.Handle)
			r.Get("/", getOrdersByUserHandler.Handle)
		})
		r.Route("/balance", func(r chi.Router) {
			r.Use(authMiddleware)
			r.Get("/", getBalanceHandler.Handle)
			r.Post("/withdraw", withdrawHandler.Handle)
		})
		r.With(authMiddleware).Get("/withdrawals", withdrawalsHandler.Handle)
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

	ctxWorker, cancelWorker := context.WithCancel(context.Background())
	go func() {
		log.Debug().Msg("⏰ Starting background worker")

		err := w.Run(ctxWorker)
		if err != nil {
			log.Error().Err(err).Msg("🟥 Worker failed")
		}

		log.Info().Msg("⬜ Worker exited")
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("⏳ Shutting down worker...")
	cancelWorker()

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
