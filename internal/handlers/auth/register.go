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

	"github.com/size12/gophermart/internal/entity"
	"github.com/size12/gophermart/internal/storage"
)

func RegisterHandler(s storage.Storage) http.HandlerFunc {
	cfg := s.GetConfig()
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

		user := entity.User{}

		err = json.Unmarshal(resBody, &user)
		if err != nil {
			http.Error(w, "wrong body: "+err.Error(), http.StatusBadRequest)
			return
		}

		user.ID, err = s.AddUser(r.Context(), user)

		if errors.Is(err, storage.ErrLoginExists) {
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
}
