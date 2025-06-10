package redis

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var (
	RedisClient *redis.Client
	clients    = make(map[*websocket.Conn]bool)
	clientsMux sync.Mutex
)


func InitRedis() {
	uri := os.Getenv("REDIS_URI")
	opt, _ := redis.ParseURL(uri)
	RedisClient = redis.NewClient(opt)
}

func StartRedisSubscriber() {
    sub := RedisClient.Subscribe(context.Background(), "game_channel")
    ch := sub.Channel()
	

    for msg := range ch {
        broadcast([]byte(msg.Payload))
    }
}

func broadcast(message []byte) {
    clientsMux.Lock()
    defer clientsMux.Unlock()

    for conn := range clients {
        err := conn.WriteMessage(websocket.TextMessage, message)
        if err != nil {
            log.Println("Broadcast error:", err)
            conn.Close()
            delete(clients, conn)
        }
    }
}