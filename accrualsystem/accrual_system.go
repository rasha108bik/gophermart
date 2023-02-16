package accrualsystem

import (
	"github.com/rasha108bik/gophermart/internal/config"
	"github.com/rasha108bik/gophermart/internal/models"
)

type AccrualSystem interface {
	GetOrderUpdates(order models.Order) (models.Order, int, error)
}

func NewAccrualSystem(cfg config.Config) AccrualSystem {
	return NewExAccrualSystem(cfg)
}
