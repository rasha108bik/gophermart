package main

import (
	"context"
	"log"
	"net/http"

	"github.com/rasha108bik/gophermart/accrualsystem"
	"github.com/rasha108bik/gophermart/internal/config"
	"github.com/rasha108bik/gophermart/internal/router"
	"github.com/rasha108bik/gophermart/internal/storage"
)

func main() {
	cfg := config.GetConfig()
	log.Printf("cfg: %#v\n", cfg)

	s, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalln("failed storage.NewStorage: ", err)
	}

	r := router.NewRouter(cfg, s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go accrualsystem.NewWorkerPool(ctx, s, accrualsystem.NewAccrualSystem(cfg))

	server := http.Server{Addr: cfg.RunAddress, Handler: r}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err.Error())
	}
}
