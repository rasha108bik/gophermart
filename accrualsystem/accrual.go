package accrualsystem

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rasha108bik/gophermart/internal/models"
	"github.com/rasha108bik/gophermart/internal/storage"
	dbErr "github.com/rasha108bik/gophermart/internal/storage/errors"
)

type Timer struct {
	Time time.Time
	*sync.RWMutex
}

type AccrualProcessor struct {
	jobs    chan models.Order
	accrual AccrualSystem
	storage storage.Storage
	timer   Timer
}

func (w *AccrualProcessor) StartWorker() {
	for {
		work := <-w.jobs

		w.timer.RLock()
		timer := w.timer.Time
		t := time.Until(timer)
		w.timer.RUnlock()

		if t.Milliseconds() > 0 {
			time.Sleep(t)
		}

		newOrderInfo, sleep, err := w.accrual.GetOrderUpdates(work)
		if err != nil {
			w.storage.PushFrontOrders(work)
			if sleep > 0 {
				w.timer.Lock()
				w.timer.Time = time.Now().Add(time.Duration(sleep) * time.Second)
				w.timer.Unlock()
			}
		}

		if newOrderInfo.Status != work.Status {
			work.Accrual = newOrderInfo.Accrual
			work.Status = newOrderInfo.Status
			w.storage.UpdateOrders(context.Background(), work)
		}
		w.storage.PushBackOrders(work)
	}
}

func NewWorkerPool(ctx context.Context, s storage.Storage, accrual AccrualSystem) {
	pool := AccrualProcessor{
		jobs:    make(chan models.Order),
		storage: s,
		accrual: accrual,
		timer: Timer{
			Time:    time.Now(),
			RWMutex: &sync.RWMutex{},
		},
	}

	go pool.StartWorker()

	for {
		job, err := s.GetOrderForUpdate()

		if errors.Is(err, dbErr.ErrEmptyQueue) {
			time.Sleep(1 * time.Second)
			continue
		}

		if err != nil {
			log.Println("failed get order for update")
			return
		}

		select {
		case pool.jobs <- job:
			fmt.Println("sent job to worker:", job)
		case <-ctx.Done():
			fmt.Println("shutdown")
			return
		}
	}

}
