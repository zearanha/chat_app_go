package main

import (
	"context"
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

	log.Println("Servidor rodando em :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}