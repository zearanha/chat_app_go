package main

import (
	"context"
	"log"
	"net/http"

	"github.com/zearanha/chat_app_go/internal/chat"
	"github.com/zearanha/chat_app_go/internal/httpapi"
	"github.com/zearanha/chat_app_go/internal/store"
)

func main() {
	ctx := context.Background()

	uri := "mongodb://localhost:27017"
	st, err := store.New(ctx, uri)
	if err != nil {
		log.Fatal("erro ao conectar no MongoDB:", err)
	}

	hub := chat.NewHub(st)
	go hub.Run()

	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, st, hub)

	log.Println("Servidor rodando em :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
