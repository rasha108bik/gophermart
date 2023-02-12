package app

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/size12/gophermart/accrualsystem"
	"github.com/size12/gophermart/internal/config"
	"github.com/size12/gophermart/internal/handlers"
	"github.com/size12/gophermart/internal/middleware"
	"github.com/size12/gophermart/internal/storage"
)

type App struct {
	Cfg config.Config
}

func NewApp(cfg config.Config) App {
	return App{Cfg: cfg}
}

func (app *App) Run() error {
	r := chi.NewRouter()
	s, err := storage.NewStorage(app.Cfg)
	if err != nil {
		log.Fatalln("Failed open storage:", err)
	}

	server := http.Server{Addr: app.Cfg.RunAddress, Handler: r}

	r.Use(middleware.RequireAuthentication(s))

	r.MethodNotAllowed(handlers.NotAllowedHandler)

	r.Post("/api/user/register", handlers.NewRegisterHandler(s))
	r.Post("/api/user/login", handlers.NewLoginHandler(s))

	r.Post("/api/user/orders", handlers.NewOrderHandler(s))
	r.Get("/api/user/orders", handlers.NewOrdersHistoryHandler(s))

	r.Post("/api/user/balance/withdraw", handlers.NewWithdrawHandler(s))
	r.Get("/api/user/withdrawals", handlers.NewWithdrawalHistoryHandler(s))

	r.Get("/api/user/balance", handlers.GetBalanceHandler(s))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go NewWorkerPool(ctx, s, accrualsystem.NewAccrualSystem(app.Cfg))

	return server.ListenAndServe()
}
