package routes

import (
	"context"
	"gameparrot_backend/redis"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	clients    = make(map[*websocket.Conn]bool)
	clientsMux sync.Mutex
	upgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	clientsMux.Lock()
	clients[conn] = true
	clientsMux.Unlock()
	log.Println("Client connected")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}
		redis.RedisClient.Publish(context.Background(), "game_channel", msg)
	}

	clientsMux.Lock()
	delete(clients, conn)
	clientsMux.Unlock()
	log.Println("Client disconnected")
}