package ws

import (
	"context"
	"gameparrot_backend/redis"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func MessageHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	redis.ClientsMux.Lock()
	redis.Clients[conn] = ""
	redis.ClientsMux.Unlock()
	log.Println("Client connected")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}
		redis.RedisClient.Publish(context.Background(), "message_channel", msg)
	}

	redis.ClientsMux.Lock()
	delete(redis.Clients, conn)
	redis.ClientsMux.Unlock()
	log.Println("Client disconnected")
}