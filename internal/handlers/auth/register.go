package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/rasha108bik/gophermart/internal/models"
	"github.com/rasha108bik/gophermart/internal/storage"
	dbErr "github.com/rasha108bik/gophermart/internal/storage/errors"
)

type Register struct {
	storage storage.Storage
}

func NewRegister(s storage.Storage) Register {
	return Register{
		storage: s,
	}
}

func (s Register) Handler(w http.ResponseWriter, r *http.Request) {
	cfg := s.storage.GetConfig()
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

	user := models.User{}

	err = json.Unmarshal(resBody, &user)
	if err != nil {
		http.Error(w, "wrong body: "+err.Error(), http.StatusBadRequest)
		return
	}

	user.ID, err = s.storage.AddUser(r.Context(), user)
	if errors.Is(err, dbErr.ErrLoginExists) {
		http.Error(w, "login already exists", http.StatusConflict)
		return
	}

	if err != nil {
		log.Println("Failed add user:", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	sessionID, err := storage.GenerateRandom()
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	h := hmac.New(sha256.New, cfg.SecretKey)
	h.Write([]byte(sessionID + fmt.Sprint(user.ID)))
	userCookie := append([]byte(sessionID), h.Sum(nil)...)
	userCookie = append(userCookie, []byte(fmt.Sprint(user.ID))...)

	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "userCookie", Value: hex.EncodeToString(userCookie), Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)
}
