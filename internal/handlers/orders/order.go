package orders

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rasha108/gophermart/internal/entity"
	"github.com/rasha108/gophermart/internal/storage"
	"github.com/theplant/luhn"
)

func OrderHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")

		if !strings.Contains(contentType, "text/plain") {
			w.Header().Set("Accept", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			http.Error(w, "wrong body: "+err.Error(), http.StatusBadRequest)
			return
		}

		orderNumber, err := strconv.Atoi(string(resBody))
		if err != nil {
			http.Error(w, "wrong order number", http.StatusBadRequest)
			return
		}

		user := r.Context().Value(entity.CtxUserKey{}).(entity.User)

		order := entity.Order{
			UserID:    user.ID,
			Number:    orderNumber,
			Status:    "NEW",
			EventTime: time.Now(),
		}

		if !luhn.Valid(order.Number) {
			http.Error(w, "wrong order number", http.StatusUnprocessableEntity)
			return
		}

		err = s.AddOrder(r.Context(), order)

		fmt.Println("added order:", order)

		if errors.Is(err, storage.ErrAlreadyLoaded) {
			w.WriteHeader(http.StatusOK)
			return
		}

		if errors.Is(err, storage.ErrLoadedByOtherUser) {
			http.Error(w, "already loaded by another user", http.StatusConflict)
			return
		}

		if err != nil {
			log.Println("Can't add new order:", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
