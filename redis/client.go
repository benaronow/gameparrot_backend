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
	Clients = make(map[*websocket.Conn]string)
	ClientsMux sync.Mutex
)


func InitRedis() {
	uri := os.Getenv("REDIS_URI")
	opt, _ := redis.ParseURL(uri)
	RedisClient = redis.NewClient(opt)
}

func StartRedisSubscriber() {
    message_sub := RedisClient.Subscribe(context.Background(), "message_channel")
    message_ch := message_sub.Channel()
	
    for msg := range message_ch {
        broadcastMessage([]byte(msg.Payload))
    }
}

func broadcastMessage(message []byte) {
    ClientsMux.Lock()
    defer ClientsMux.Unlock()

    for conn := range Clients {
        err := conn.WriteMessage(websocket.TextMessage, message)
        if err != nil {
            log.Println("Broadcast error:", err)
            conn.Close()
            delete(Clients, conn)
        }
    }
}