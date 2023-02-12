package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/rasha108bik/gophermart/internal/models"
)

func RequireAuthentication(secretKey []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			path := r.URL.Path
			if path == "/api/user/register" || path == "/api/user/login" {
				next.ServeHTTP(w, r)
				return
			}

			userCookie, err := r.Cookie("userCookie")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			user := models.User{}

			cookie, err := hex.DecodeString(userCookie.Value)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			sign := append([]byte{cookie[8]}, cookie[9:40]...)
			data := append(cookie[:8], cookie[40:]...)

			h := hmac.New(sha256.New, secretKey)
			h.Write(data)
			s := h.Sum(nil)

			if !hmac.Equal(sign, s) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			user.ID, err = strconv.Atoi(string(cookie[40:]))
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), models.CtxUserKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
