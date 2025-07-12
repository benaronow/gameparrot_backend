package ws

import (
	"encoding/json"
	"fmt"
	"gameparrot_backend/models"
	"gameparrot_backend/redis"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func HandleWSConnection(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
	}

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
	log.Println("Message client connected")

	key := fmt.Sprintf("user:%s:online", userID)
    err = redis.RedisClient.Set(ctx, key, "1", time.Minute).Err()
    if err != nil {
        log.Println("Redis set online error:", err)
    }
    redis.RedisClient.Publish(ctx, "status_channel", "")

	go func() {
		defer func() {
			redis.ClientsMux.Lock()
			delete(redis.Clients, conn)
			redis.ClientsMux.Unlock()
			log.Println("Message client disconnected")

			err := redis.RedisClient.Del(ctx, key).Err()
            if err != nil {
                log.Println("Redis offline error:", err)
            }

            redis.RedisClient.Publish(ctx, "status_channel", "")
            conn.Close()
		}()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket read error:", err)
				break
			}
	
			var updateJson models.Update
			if err := json.Unmarshal(msg, &updateJson); err != nil {
				log.Println("JSON decode error:", err)
				continue
			}

			switch updateJson.Type {
				case models.UpdateTypeMessage:
					handleMessageUpdate(updateJson);
					continue;
				case models.UpdateTypeFriend:
					continue;
				default:
					continue;
			}
		}
	}()
}