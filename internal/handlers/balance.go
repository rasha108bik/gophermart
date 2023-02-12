package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/size12/gophermart/internal/entity"
	"github.com/size12/gophermart/internal/storage"
)

func GetBalanceHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		userID := r.Context().Value(entity.CtxUserKey{}).(entity.User).ID

		user, err := s.GetUser(ctx, "id", fmt.Sprint(userID))

		if err != nil {
			log.Println("Failed fetch balance:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		balance := entity.Balance{
			Current:   user.Balance,
			Withdrawn: user.Withdrawn,
		}

		b, err := json.Marshal(balance)

		if err != nil {
			log.Println("Failed marshalling balance:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}
}
