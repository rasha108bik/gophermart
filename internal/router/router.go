package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/rasha108bik/gophermart/internal/config"
	"github.com/rasha108bik/gophermart/internal/handlers"
	"github.com/rasha108bik/gophermart/internal/handlers/auth"
	"github.com/rasha108bik/gophermart/internal/handlers/balance"
	"github.com/rasha108bik/gophermart/internal/handlers/orders"
	"github.com/rasha108bik/gophermart/internal/handlers/withdraw"
	"github.com/rasha108bik/gophermart/internal/middleware"
	"github.com/rasha108bik/gophermart/internal/storage"
)

func NewRouter(cfg config.Config, s storage.Storage) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuthentication(cfg.SecretKey))
	r.MethodNotAllowed(handlers.NotAllowedHandler)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", auth.NewRegister(s).Handler)
		r.Post("/login", auth.NewLogin(s).Handler)

		r.Route("/orders", func(r chi.Router) {
			r.Post("/", orders.NewOrder(s).Handler)
			r.Get("/", orders.NewHistory(s).Handler)
		})

		r.Post("/balance/withdraw", withdraw.NewWithdraw(s).Handler)
		r.Get("/balance", balance.NewBalance(s).Handler)
		r.Get("/withdrawals", withdraw.NewWithdrawalHistory(s).Handler)
	})

	return r
}
