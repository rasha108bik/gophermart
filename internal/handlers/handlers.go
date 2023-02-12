package handlers

import (
	"net/http"

	"github.com/size12/gophermart/internal/handlers/auth"
	"github.com/size12/gophermart/internal/handlers/orders"
	"github.com/size12/gophermart/internal/handlers/withdraw"
	"github.com/size12/gophermart/internal/storage"
)

func NewLoginHandler(s storage.Storage) http.HandlerFunc {
	return auth.LoginHandler(s)
}

func NewRegisterHandler(s storage.Storage) http.HandlerFunc {
	return auth.RegisterHandler(s)
}

func NewWithdrawHandler(s storage.Storage) http.HandlerFunc {
	return withdraw.WithdrawHandler(s)
}

func NewWithdrawalHistoryHandler(s storage.Storage) http.HandlerFunc {
	return withdraw.WithdrawalHistoryHandler(s)
}

func NewOrderHandler(s storage.Storage) http.HandlerFunc {
	return orders.OrderHandler(s)
}

func NewOrdersHistoryHandler(s storage.Storage) http.HandlerFunc {
	return orders.HistoryHandler(s)
}
