package ws

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"gameparrot_backend/redis"
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("WS upgrade error:", err)
        return
    }

    _, msg, err := conn.ReadMessage()
    if err != nil {
        log.Println("Read user ID error:", err)
        conn.Close()
        return
    }

    userID := string(msg)
    redis.ClientsMux.Lock()
    redis.StatusClients[conn] = userID
    redis.ClientsMux.Unlock()

    key := fmt.Sprintf("user:%s:online", userID)
    err = redis.RedisClient.Set(ctx, key, "1", time.Minute).Err()
    if err != nil {
        log.Println("Redis set online error:", err)
    }

    redis.RedisClient.Publish(ctx, "status_channel", "")

    go func() {
        defer func() {
            redis.ClientsMux.Lock()
            delete(redis.StatusClients, conn)
            redis.ClientsMux.Unlock()

            err := redis.RedisClient.Del(ctx, key).Err()
            if err != nil {
                log.Println("Redis offline error:", err)
            }

            redis.RedisClient.Publish(context.Background(), "status_channel", "")
            conn.Close()
        }()

        for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket read error:", err)
				break
			}
			redis.RedisClient.Publish(context.Background(), "status_channel", "")
		}
    }()
}