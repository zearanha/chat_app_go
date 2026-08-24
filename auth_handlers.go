package main

import (
	"context"
	"encoding/json"
	"net/http"
)

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

func registerHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "corpo inválido", http.StatusBadRequest)
			return
		}
		if req.Username == "" || req.Password == "" {
			http.Error(w, "username e password são obrigatórios", http.StatusBadRequest)
			return
		}

		existing, err := store.FindUserByUsername(context.Background(), req.Username)
		if err != nil {
			http.Error(w, "erro interno", http.StatusInternalServerError)
			return
		}
		if existing != nil {
			http.Error(w, "usuário já existe", http.StatusConflict)
			return
		}

		hash, err := hashPassword(req.Password)
		if err != nil {
			http.Error(w, "erro interno", http.StatusInternalServerError)
			return
		}

		if _, err := store.CreateUser(context.Background(), req.Username, hash); err != nil {
			http.Error(w, "erro ao criar usuário", http.StatusInternalServerError)
			return
		}

		token, err := generateToken(req.Username)
		if err != nil {
			http.Error(w, "erro ao gerar token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(authResponse{Token: token})
	}
}

func loginHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "corpo inválido", http.StatusBadRequest)
			return
		}

		user, err := store.FindUserByUsername(context.Background(), req.Username)
		if err != nil || user == nil {
			http.Error(w, "credenciais inválidas", http.StatusUnauthorized)
			return
		}

		if !checkPassword(req.Password, user.PasswordHash) {
			http.Error(w, "credenciais inválidas", http.StatusUnauthorized)
			return
		}

		token, err := generateToken(req.Username)
		if err != nil {
			http.Error(w, "erro ao gerar token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(authResponse{Token: token})
	}
}