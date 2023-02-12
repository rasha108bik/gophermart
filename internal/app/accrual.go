package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/size12/gophermart/accrualsystem"
	"github.com/size12/gophermart/internal/entity"
	"github.com/size12/gophermart/internal/storage"
)

type Timer struct {
	Time time.Time
	*sync.RWMutex
}

type WorkerPool struct {
	jobs    chan entity.Order
	accrual accrualsystem.AccrualSystem
	storage storage.Storage
	timer   Timer
}

func (w *WorkerPool) StartWorker() {
	go func() {
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
				log.Println("Failed get update order info:", err)
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
			} else {
				w.storage.PushBackOrders(work)
			}
		}
	}()
}

func NewWorkerPool(ctx context.Context, s storage.Storage, accrual accrualsystem.AccrualSystem) {
	pool := WorkerPool{
		jobs:    make(chan entity.Order),
		storage: s,
		accrual: accrual,
		timer: Timer{
			Time:    time.Now(),
			RWMutex: &sync.RWMutex{},
		},
	}

	pool.StartWorker()

	for {
		job, err := s.GetOrderForUpdate()

		if errors.Is(err, storage.ErrEmptyQueue) {
			time.Sleep(1 * time.Second)
			continue
		}

		if err != nil {
			log.Println("Failed get order for update")
			return
		}

		select {
		case pool.jobs <- job:
			fmt.Println("Sent job to worker:", job)
		case <-ctx.Done():
			fmt.Println("Shutdown")
			return
		}
	}

}
