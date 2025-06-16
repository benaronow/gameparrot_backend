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
	Clients = make(map[*websocket.Conn]string)
	ClientsMux sync.Mutex
    ctx = context.Background()
)


func InitRedis() {
	uri := os.Getenv("REDIS_URI")
	opt, _ := redis.ParseURL(uri)
	RedisClient = redis.NewClient(opt)
}

func StartRedisSubscriber() {
    sub := RedisClient.Subscribe(ctx, "message_channel", "status_channel")
    ch := sub.Channel()

    fmt.Println("Redis subscriber listening on message and status channels")

    for msg := range ch {
        switch msg.Channel {
        case "message_channel":
            broadcastMessage([]byte(msg.Payload))
        case "status_channel":
            fmt.Println("Received status update")
            broadcastStatus()
        default:
            fmt.Println("Received message on unknown channel:", msg.Channel)
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

func broadcastMessage(message []byte) {
    ClientsMux.Lock()
    defer ClientsMux.Unlock()

    var msgJson models.Message
	if err := json.Unmarshal(message, &msgJson); err != nil {
		log.Println("JSON decode error:", err)
    }

    for conn := range Clients {
        if (Clients[conn] == msgJson.To) {
            err := conn.WriteMessage(websocket.TextMessage, message)
            if err != nil {
                log.Println("Broadcast error:", err)
                conn.Close()
                delete(Clients, conn)
            }
        }
    }
}

func broadcastStatus() {
    ClientsMux.Lock()
    defer ClientsMux.Unlock()

    for conn := range Clients {
        status := getStatus()

        statusMsgJson := models.Update{ Type: "status", Status: status}
        statusMsg, msgErr := json.Marshal(statusMsgJson)

        if err := conn.WriteMessage(websocket.TextMessage, statusMsg); err != nil || msgErr != nil {
            log.Println("WriteMessage error:", err)
            conn.Close()
            delete(Clients, conn)
        }
    }
} 

func getStatus() []models.User {
    cursor, err := mongo.UserCollection.Find(ctx, bson.M{})
    if err != nil {
        log.Println("Mongo find error:", err)
        return nil
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

        users = append(users, abbrevUser(user))
    }

    return users
}

func abbrevUser(user models.User) models.User {
    return models.User{
        UID: user.UID,
        Email: user.Email,
        Online: user.Online,
    }
}