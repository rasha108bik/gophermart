package main

import (
	"log"

	"github.com/rasha108/gophermart/internal/app"
	"github.com/rasha108/gophermart/internal/config"
)

func main() {
	cfg := config.GetConfig()
	service := app.NewApp(cfg)
	log.Fatal(service.Run())
}
