package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "token ausente", http.StatusUnauthorized)
		return
	}

	claims, err := validateToken(tokenString)
	if err != nil {
		http.Error(w, "token inválido", http.StatusUnauthorized)
		return
	}

	room := r.URL.Query().Get("room")
	if room == "" {
		room = "geral" // sala padrão
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		username: claims.Username,
		room:     room,
	}
	client.hub.register <- client

	go sendHistory(hub, client)

	go client.writePump()
	go client.readPump()
}

func sendHistory(hub *Hub, client *Client) {
	messages, err := hub.store.RecentMessages(context.Background(), client.room, 20)
	if err != nil {
		log.Println("erro ao buscar histórico:", err)
		return
	}
	for _, msg := range messages {
		payload, err := json.Marshal(msg)
		if err != nil {
			continue
		}
		client.send <- payload
	}
}