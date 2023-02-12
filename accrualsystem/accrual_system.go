package accrualsystem

import (
	"github.com/size12/gophermart/internal/config"
	"github.com/size12/gophermart/internal/entity"
)

type AccrualSystem interface {
	GetOrderUpdates(order entity.Order) (entity.Order, int, error)
}

func NewAccrualSystem(cfg config.Config) AccrualSystem {
	return NewExAccrualSystem(cfg)
}
