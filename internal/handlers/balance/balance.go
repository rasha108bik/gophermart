package balance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rasha108bik/gophermart/internal/models"
	"github.com/rasha108bik/gophermart/internal/storage"
)

type Balance struct {
	storage storage.Storage
}

func NewBalance(s storage.Storage) Balance {
	return Balance{
		storage: s,
	}
}

func (s Balance) Handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	userID := r.Context().Value(models.CtxUserKey{}).(models.User).ID

	user, err := s.storage.GetUser(ctx, "id", fmt.Sprint(userID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	balance := models.Balance{
		Current:   user.Balance,
		Withdrawn: user.Withdrawn,
	}

	b, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
