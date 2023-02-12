package storage

import (
	"sync"

	"github.com/size12/gophermart/internal/entity"
)

type Queue interface {
	PushFrontOrders(orders ...entity.Order) error
	PushBackOrders(orders ...entity.Order) error
	GetOrder() (entity.Order, error)
}

type SliceQueue struct {
	Orders []entity.Order
	*sync.Mutex
}

func NewSliceQueue() *SliceQueue {
	return &SliceQueue{
		Mutex: &sync.Mutex{},
	}
}

func (q *SliceQueue) PushFrontOrders(orders ...entity.Order) error {
	q.Lock()
	defer q.Unlock()
	q.Orders = append(orders, q.Orders...)
	return nil
}

func (q *SliceQueue) PushBackOrders(orders ...entity.Order) error {
	q.Lock()
	defer q.Unlock()
	q.Orders = append(q.Orders, orders...)
	return nil
}

func (q *SliceQueue) GetOrder() (entity.Order, error) {
	q.Lock()
	defer q.Unlock()
	if len(q.Orders) > 0 {
		order := q.Orders[0]
		q.Orders = q.Orders[1:]
		return order, nil
	}
	return entity.Order{}, ErrEmptyQueue
}
