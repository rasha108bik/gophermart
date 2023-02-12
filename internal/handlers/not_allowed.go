package handlers

import "net/http"

func NotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "wrong method", http.StatusMethodNotAllowed)
}
