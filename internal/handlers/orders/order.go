package orders

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rasha108bik/gophermart/internal/models"
	"github.com/rasha108bik/gophermart/internal/storage"
	dbErr "github.com/rasha108bik/gophermart/internal/storage/errors"

	"github.com/theplant/luhn"
)

type Order struct {
	storage storage.Storage
}

func NewOrder(s storage.Storage) Order {
	return Order{
		storage: s,
	}
}

func (s Order) Handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		w.Header().Set("Accept", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderNumber, err := strconv.Atoi(string(resBody))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := r.Context().Value(models.CtxUserKey{}).(models.User)

	order := models.Order{
		UserID:    user.ID,
		Number:    orderNumber,
		Status:    "NEW",
		EventTime: time.Now(),
	}

	if !luhn.Valid(order.Number) {
		http.Error(w, "wrong order number", http.StatusUnprocessableEntity)
		return
	}

	fmt.Println("added order: ", order)
	err = s.storage.AddOrder(r.Context(), order)
	if errors.Is(err, dbErr.ErrAlreadyLoaded) {
		w.WriteHeader(http.StatusOK)
		return
	}

	if errors.Is(err, dbErr.ErrLoadedByOtherUser) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
