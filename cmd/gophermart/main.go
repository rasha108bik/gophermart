package main

import (
	"log"

	"github.com/size12/gophermart/internal/app"
	"github.com/size12/gophermart/internal/config"
)

func main() {
	cfg := config.GetConfig()
	service := app.NewApp(cfg)
	log.Fatal(service.Run())
}
