package withdraw

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/rasha108bik/gophermart/internal/models"
	"github.com/rasha108bik/gophermart/internal/storage"
	dbErr "github.com/rasha108bik/gophermart/internal/storage/errors"

	"github.com/theplant/luhn"
)

type Withdraw struct {
	storage storage.Storage
}

func NewWithdraw(s storage.Storage) Withdraw {
	return Withdraw{
		storage: s,
	}
}

func (s Withdraw) Handler(w http.ResponseWriter, r *http.Request) {
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

	withdrawal := models.Withdraw{}

	err = json.Unmarshal(resBody, &withdrawal)
	if err != nil {
		http.Error(w, "wrong body: "+err.Error(), http.StatusBadRequest)
		return
	}

	user := r.Context().Value(models.CtxUserKey{}).(models.User)

	if !luhn.Valid(withdrawal.Order) {
		http.Error(w, "wrong order number", http.StatusUnprocessableEntity)
		return
	}

	err = s.storage.Withdraw(r.Context(), user, withdrawal)

	if errors.Is(err, dbErr.ErrNoMoney) {
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
