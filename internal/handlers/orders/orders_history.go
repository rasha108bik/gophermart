package orders

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/size12/gophermart/internal/entity"
	"github.com/size12/gophermart/internal/storage"
)

func HistoryHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(entity.CtxUserKey{}).(entity.User)

		withdrawals, err := s.OrdersHistory(r.Context(), user)
		if err != nil {
			log.Println("Can't get orders history:", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		if len(withdrawals) == 0 {
			http.Error(w, "no content", http.StatusNoContent)
			return
		}

		b, err := json.Marshal(withdrawals)
		if err != nil {
			log.Println("Can't marshal orders history:", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}
}
