package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/dr3vv5/go_final_project/pkg/auth"
)

type SigninRequest struct {
	Password string `json:"password"`
}

type SigninResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
	correctPassword := os.Getenv("TODO_PASSWORD")

	if correctPassword == "" {
		writeJSONError(w, "TODO_PASSWORD is not set", http.StatusInternalServerError)
		return
	}

	var req SigninRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Password != correctPassword {
		writeJSONError(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(req.Password)
	if err != nil {
		writeJSONError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   28800,
	}
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	resp := SigninResponse{Token: token}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return
	}
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correctPassword := os.Getenv("TODO_PASSWORD")

		if correctPassword == "" {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		isValid, validateErr := auth.ValidateToken(cookie.Value, correctPassword)

		if !isValid || validateErr != nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		next(w, r)
	})
}
