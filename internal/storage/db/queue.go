package db

import (
	"sync"

	"github.com/rasha108bik/gophermart/internal/models"
	dbErr "github.com/rasha108bik/gophermart/internal/storage/errors"
)

type Queue interface {
	PushFrontOrders(orders ...models.Order) error
	PushBackOrders(orders ...models.Order) error
	GetOrder() (models.Order, error)
}

type SliceQueue struct {
	Orders []models.Order
	*sync.Mutex
}

func NewSliceQueue() *SliceQueue {
	return &SliceQueue{
		Mutex: &sync.Mutex{},
	}
}

func (q *SliceQueue) PushFrontOrders(orders ...models.Order) error {
	q.Lock()
	defer q.Unlock()
	q.Orders = append(orders, q.Orders...)
	return nil
}

func (q *SliceQueue) PushBackOrders(orders ...models.Order) error {
	q.Lock()
	defer q.Unlock()
	q.Orders = append(q.Orders, orders...)
	return nil
}

func (q *SliceQueue) GetOrder() (models.Order, error) {
	q.Lock()
	defer q.Unlock()
	if len(q.Orders) > 0 {
		order := q.Orders[0]
		q.Orders = q.Orders[1:]
		return order, nil
	}
	return models.Order{}, dbErr.ErrEmptyQueue
}
