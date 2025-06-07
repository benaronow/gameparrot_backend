package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
    ctx         = context.Background()
    redisClient *redis.Client
    upgrader    = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
    }
    clients     = make(map[*websocket.Conn]bool)
    clientsLock sync.Mutex
)

func main() {
	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
	
	redisURI := os.Getenv("REDIS_URI")
	opt, _ := redis.ParseURL(redisURI)
    redisClient = redis.NewClient(opt)

    // Start listening to Redis pub/sub
    go startRedisSubscriber()

    http.HandleFunc("/ws", wsHandler)
    fmt.Println("WebSocket server running at :8080/ws")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("WebSocket upgrade error:", err)
        return
    }
    defer conn.Close()

    // Register client
    clientsLock.Lock()
    clients[conn] = true
    clientsLock.Unlock()
    log.Println("Client connected")

    // Handle incoming messages
    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            log.Println("WebSocket read error:", err)
            break
        }

        // Publish to Redis
        err = redisClient.Publish(ctx, "game_channel", msg).Err()
        if err != nil {
            log.Println("Redis publish error:", err)
        }
    }

    // Cleanup on disconnect
    clientsLock.Lock()
    delete(clients, conn)
    clientsLock.Unlock()
    log.Println("Client disconnected")
}

func startRedisSubscriber() {
    sub := redisClient.Subscribe(ctx, "game_channel")
    ch := sub.Channel()
	

    for msg := range ch {
        broadcast([]byte(msg.Payload))
    }
}

func broadcast(message []byte) {
    clientsLock.Lock()
    defer clientsLock.Unlock()

    for conn := range clients {
        err := conn.WriteMessage(websocket.TextMessage, message)
        if err != nil {
            log.Println("Broadcast error:", err)
            conn.Close()
            delete(clients, conn)
        }
    }
}
