package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/zearanha/chat_app_go/internal/auth"
	"github.com/zearanha/chat_app_go/internal/chat"
	"github.com/zearanha/chat_app_go/internal/store"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveWs(hub *chat.Hub, store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.URL.Query().Get("token")
		if tokenString == "" {
			http.Error(w, "token ausente", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "token invalido", http.StatusUnauthorized)
			return
		}

		room := r.URL.Query().Get("room")
		if room == "" {
			room = "geral"
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		client := chat.NewClient(hub, conn, claims.Username, room)
		client.Start()

		go sendHistory(store, client)
	}
}

func sendHistory(store *store.Store, client *chat.Client) {
	messages, err := store.RecentMessages(context.Background(), client.Room(), 20)
	if err != nil {
		log.Println("erro ao buscar historico:", err)
		return
	}
	for _, msg := range messages {
		payload, err := json.Marshal(msg)
		if err != nil {
			continue
		}
		client.Send(payload)
	}
}
