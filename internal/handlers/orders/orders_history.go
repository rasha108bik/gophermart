package orders

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rasha108bik/gophermart/internal/models"
	"github.com/rasha108bik/gophermart/internal/storage"
)

type History struct {
	storage storage.Storage
}

func NewHistory(s storage.Storage) History {
	return History{
		storage: s,
	}
}

func (s History) Handler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(models.CtxUserKey{}).(models.User)

	withdrawals, err := s.storage.OrdersHistory(r.Context(), user)
	if err != nil {
		log.Println("Can't get orders history:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		http.Error(w, "no content", http.StatusNoContent)
		return
	}

	b, err := json.Marshal(withdrawals)
	if err != nil {
		log.Println("Can't marshal orders history:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
