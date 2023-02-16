package withdraw

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rasha108bik/gophermart/internal/models"
	"github.com/rasha108bik/gophermart/internal/storage"
)

type WithdrawalHistory struct {
	storage storage.Storage
}

func NewWithdrawalHistory(s storage.Storage) WithdrawalHistory {
	return WithdrawalHistory{
		storage: s,
	}
}

func (s WithdrawalHistory) Handler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(models.CtxUserKey{}).(models.User)

	withdrawals, err := s.storage.WithdrawalHistory(r.Context(), user)
	if err != nil {
		log.Println("Can't get withdrawal history:", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		http.Error(w, "no content", http.StatusNoContent)
		return
	}

	b, err := json.Marshal(withdrawals)
	if err != nil {
		log.Println("Can't marshal withdrawal history:", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
