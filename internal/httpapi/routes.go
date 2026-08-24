package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/zearanha/chat_app_go/internal/chat"
	"github.com/zearanha/chat_app_go/internal/store"
)

func RegisterRoutes(mux *http.ServeMux, store *store.Store, hub *chat.Hub) {
	mux.HandleFunc("/register", registerHandler(store))
	mux.HandleFunc("/login", loginHandler(store))
	mux.HandleFunc("/ws", serveWs(hub, store))
	mux.HandleFunc("/rooms", listRoomsHandler(store))
}

func listRoomsHandler(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rooms, err := store.ListRooms(context.Background())
		if err != nil {
			http.Error(w, "erro ao listar salas", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rooms)
	}
}
