package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"gameparrot_backend/models"
	"gameparrot_backend/mongo"
	"log"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	RedisClient *redis.Client
	MessageClients = make(map[*websocket.Conn]string)
    StatusClients = make(map[*websocket.Conn]string)
	ClientsMux sync.Mutex
)


func InitRedis() {
	uri := os.Getenv("REDIS_URI")
	opt, _ := redis.ParseURL(uri)
	RedisClient = redis.NewClient(opt)
}

func StartRedisSubscriber() {
    sub := RedisClient.Subscribe(context.Background(), "message_channel", "status_channel")
    ch := sub.Channel()

    fmt.Println("Redis subscriber listening on message_channel and status_channel")

    for msg := range ch {
        switch msg.Channel {
        case "message_channel":
            broadcastMessage([]byte(msg.Payload))
        case "status_channel":
            fmt.Println("Received status update:", msg.Payload)
            broadcastStatus()
        default:
            fmt.Println("Received message on unknown channel:", msg.Channel)
        }
    }
}

func broadcastMessage(message []byte) {
    ClientsMux.Lock()
    defer ClientsMux.Unlock()

    for conn := range MessageClients {
        err := conn.WriteMessage(websocket.TextMessage, message)
        if err != nil {
            log.Println("Broadcast error:", err)
            conn.Close()
            delete(MessageClients, conn)
        }
    }
}

func StartStatusRebroadcast() {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			log.Println("Rebroadcasting user statuses...")
			broadcastStatus()
		}
	}()
}

func broadcastStatus() {
    ctx := context.Background()

    cursor, err := mongo.UserCollection.Find(ctx, bson.M{})
    if err != nil {
        log.Println("Mongo find error:", err)
        return
    }
    defer cursor.Close(ctx)

    var users []models.User
    for cursor.Next(ctx) {
        var user models.User
        if err := cursor.Decode(&user); err != nil {
            log.Println("Mongo decode user error:", err)
            continue
        }

        key := fmt.Sprintf("user:%s:online", user.UID)
        online, err := RedisClient.Exists(ctx, key).Result()
        if err != nil {
            log.Printf("Redis Exists error for %s: %v\n", user.UID, err)
            online = 0
        }
        user.Online = (online == 1)

        users = append(users, user)
    }

    data, err := json.Marshal(users)
    if err != nil {
        log.Println("JSON marshal error:", err)
        return
    }

    ClientsMux.Lock()
    defer ClientsMux.Unlock()

    for conn := range StatusClients {
        if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
            log.Println("WriteMessage error:", err)
            conn.Close()
            delete(StatusClients, conn)
        }
    }
}