package storage

import (
	"context"
	"math/rand"

	"github.com/rasha108bik/gophermart/internal/config"
	"github.com/rasha108bik/gophermart/internal/models"
	"github.com/rasha108bik/gophermart/internal/storage/db"
)

type Storage interface {
	GetUser(ctx context.Context, search, value string) (models.User, error)
	AddUser(ctx context.Context, user models.User) (int, error)
	Withdraw(ctx context.Context, user models.User, order models.Withdraw) error
	WithdrawalHistory(ctx context.Context, user models.User) ([]models.Withdraw, error)
	AddOrder(ctx context.Context, order models.Order) error
	OrdersHistory(ctx context.Context, user models.User) ([]models.Order, error)
	GetOrdersForUpdate(ctx context.Context) ([]models.Order, error)
	GetOrderForUpdate() (models.Order, error)
	UpdateOrders(ctx context.Context, orders ...models.Order) error
	GetConfig() config.Config
	PushFrontOrders(orders ...models.Order) error
	PushBackOrders(orders ...models.Order) error
}

func NewStorage(cfg config.Config) (Storage, error) {
	return db.NewDBStorage(cfg)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenerateRandom() (string, error) {
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b), nil
}
