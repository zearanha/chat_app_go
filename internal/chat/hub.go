package chat

import (
	"context"
	"encoding/json"
	"log"

	"github.com/zearanha/chat_app_go/internal/store"
)

type broadcastMessage struct {
	username string
	room     string
	content  string
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan broadcastMessage
	register   chan *Client
	unregister chan *Client
	store      *store.Store
}

func NewHub(store *store.Store) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan broadcastMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		store:      store,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case bm := <-h.broadcast:
			msg, err := h.store.SaveMessage(context.Background(), bm.username, bm.room, bm.content)
			if err != nil {
				log.Println("erro ao salvar mensagem:", err)
				continue
			}

			payload, err := json.Marshal(msg)
			if err != nil {
				log.Println("erro ao serializar mensagem:", err)
				continue
			}

			for client := range h.clients {
				if client.room != bm.room {
					continue
				}
				select {
				case client.send <- payload:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
