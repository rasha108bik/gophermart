package withdraw

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/rasha108/gophermart/internal/entity"
	"github.com/rasha108/gophermart/internal/storage"
	"github.com/theplant/luhn"
)

func WithdrawHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")

		if !strings.Contains(contentType, "application/json") {
			w.Header().Set("Accept", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			http.Error(w, "wrong body: "+err.Error(), http.StatusBadRequest)
			return
		}

		withdrawal := entity.Withdraw{}

		err = json.Unmarshal(resBody, &withdrawal)
		if err != nil {
			http.Error(w, "wrong body: "+err.Error(), http.StatusBadRequest)
			return
		}

		user := r.Context().Value(entity.CtxUserKey{}).(entity.User)

		if !luhn.Valid(withdrawal.Order) {
			http.Error(w, "wrong order number", http.StatusUnprocessableEntity)
			return
		}

		err = s.Withdraw(r.Context(), user, withdrawal)

		if errors.Is(err, storage.ErrNoMoney) {
			http.Error(w, "Insufficient balance", http.StatusUnprocessableEntity)
			return
		}

		if err != nil {
			log.Println("Can't withdraw money:", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
