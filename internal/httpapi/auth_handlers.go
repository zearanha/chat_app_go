package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/zearanha/chat_app_go/internal/auth"
	"github.com/zearanha/chat_app_go/internal/store"
)

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

func registerHandler(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "corpo invalido", http.StatusBadRequest)
			return
		}
		if req.Username == "" || req.Password == "" {
			http.Error(w, "username e password sao obrigatorios", http.StatusBadRequest)
			return
		}

		existing, err := store.FindUserByUsername(context.Background(), req.Username)
		if err != nil {
			http.Error(w, "erro interno", http.StatusInternalServerError)
			return
		}
		if existing != nil {
			http.Error(w, "usuario ja existe", http.StatusConflict)
			return
		}

		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			http.Error(w, "erro interno", http.StatusInternalServerError)
			return
		}

		if _, err := store.CreateUser(context.Background(), req.Username, hash); err != nil {
			http.Error(w, "erro ao criar usuario", http.StatusInternalServerError)
			return
		}

		token, err := auth.GenerateToken(req.Username)
		if err != nil {
			http.Error(w, "erro ao gerar token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(authResponse{Token: token})
	}
}

func loginHandler(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "corpo invalido", http.StatusBadRequest)
			return
		}

		user, err := store.FindUserByUsername(context.Background(), req.Username)
		if err != nil || user == nil {
			http.Error(w, "credenciais invalidas", http.StatusUnauthorized)
			return
		}

		if !auth.CheckPassword(req.Password, user.PasswordHash) {
			http.Error(w, "credenciais invalidas", http.StatusUnauthorized)
			return
		}

		token, err := auth.GenerateToken(req.Username)
		if err != nil {
			http.Error(w, "erro ao gerar token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(authResponse{Token: token})
	}
}
