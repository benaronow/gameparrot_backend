package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"gameparrot_backend/models"
	"gameparrot_backend/mongo"
	"gameparrot_backend/redis"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

type User struct {
	ID     string `json:"uid" bson:"_id"`
	Email  string `json:"email"`
	Online bool   `json:"online"`
}

// WebSocket handler
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("WS upgrade error:", err)
        return
    }

    // Read first message which should be user ID
    _, msg, err := conn.ReadMessage()
    if err != nil {
        log.Println("Read user ID error:", err)
        conn.Close()
        return
    }
    userID := string(msg)

    redis.ClientsMux.Lock()
    redis.Clients[conn] = userID
    redis.ClientsMux.Unlock()

    ctx := context.Background()
    // Mark user online in Redis (set with expiration 30 minutes)
    key := fmt.Sprintf("user:%s:online", userID)
    err = redis.RedisClient.Set(ctx, key, "1", 0).Err() // 0 = no expiration, or add time.Duration if you want expiry
    if err != nil {
        log.Println("Redis set online error:", err)
    }

    broadcastStatus() // broadcast updated status

    // Listen for disconnect and messages (optional)
    go func() {
        defer func() {
            // Remove from clients map
            redis.ClientsMux.Lock()
            delete(redis.Clients, conn)
            redis.ClientsMux.Unlock()

            // Mark user offline
            err := redis.RedisClient.Del(ctx, key).Err()
            if err != nil {
                log.Println("Redis del offline error:", err)
            }

            broadcastStatus() // broadcast updated status
            conn.Close()
        }()

        for {
            if _, _, err := conn.ReadMessage(); err != nil {
                break
            }
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
        online, err := redis.RedisClient.Exists(ctx, key).Result()
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

    redis.ClientsMux.Lock()
    defer redis.ClientsMux.Unlock()
    for conn := range redis.Clients {
        if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
            log.Println("WriteMessage error:", err)
            conn.Close()
            delete(redis.Clients, conn)
        }
    }
}