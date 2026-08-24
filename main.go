package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()

	uri := "mongodb://localhost:27017"
	store, err := newStore(ctx, uri)
	if err != nil {
		log.Fatal("erro ao conectar no MongoDB:", err)
	}

	hub := newHub(store)
	go hub.run()

	http.HandleFunc("/register", registerHandler(store))
	http.HandleFunc("/login", loginHandler(store))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	http.HandleFunc("/rooms", func(w http.ResponseWriter, r *http.Request) {
	rooms, err := store.ListRooms(context.Background())
	if err != nil {
		http.Error(w, "erro ao listar salas", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
})

	log.Println("Servidor rodando em :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}