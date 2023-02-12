package accrualsystem

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/rasha108bik/gophermart/internal/config"
	"github.com/rasha108bik/gophermart/internal/models"
)

type ExAccrualSystem struct {
	BaseURL string
}

func NewExAccrualSystem(cfg config.Config) *ExAccrualSystem {
	return &ExAccrualSystem{BaseURL: cfg.AccrualSystemAddress}
}

func (s *ExAccrualSystem) GetOrderUpdates(order models.Order) (models.Order, int, error) {
	sleep := 0

	path := "/api/orders/"
	url := fmt.Sprintf("%s%s%v", s.BaseURL, path, order.Number)

	r, err := http.Get(url)
	if err != nil {
		log.Println("Can't get order updates from external API:", err)
		return order, sleep, err
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		log.Println("Can't read response body:", err)
		return order, sleep, err
	}

	if r.StatusCode == http.StatusNoContent {
		return order, 0, nil
	}

	if r.StatusCode == http.StatusTooManyRequests {
		res, err := strconv.Atoi(r.Header.Get("Retry-After"))
		if err != nil {
			return order, 0, err
		}
		return order, res, err
	}

	fmt.Println(r.StatusCode)
	fmt.Println(string(body))

	err = json.Unmarshal(body, &order)

	if err != nil {
		return order, sleep, err
	}

	return order, sleep, nil
}
