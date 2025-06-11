package ws

import (
	"encoding/json"
	"fmt"
	"gameparrot_backend/models"
	"gameparrot_backend/redis"
	"log"
	"net/http"
	"time"
)

func MessageHandler(w http.ResponseWriter, r *http.Request) {
	conn := getConn(w, r)

	_, msg, err := conn.ReadMessage()
    if err != nil {
        log.Println("Read user ID error:", err)
        conn.Close()
        return
    }
    userID := string(msg)

	redis.ClientsMux.Lock()
	redis.MessageClients[conn] = userID
	redis.ClientsMux.Unlock()
	log.Println("Message client connected")

	go func() {
		defer func() {
			redis.ClientsMux.Lock()
			delete(redis.MessageClients, conn)
			redis.ClientsMux.Unlock()
			log.Println("Message client disconnected")
			conn.Close()
		}()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket read error:", err)
				break
			}
	
			var msgJson models.Message
			if err := json.Unmarshal(msg, &msgJson); err != nil {
				log.Println("JSON decode error:", err)
				continue
			}
	
			key := fmt.Sprintf("user:%s:online", msgJson.From)
			err = redis.RedisClient.Set(ctx, key, "1", time.Minute).Err()
			if err != nil {
				log.Println("Redis set online error:", err)
			} else {
				redis.RedisClient.Publish(ctx, "status_channel", "")
				redis.RedisClient.Publish(ctx, "message_channel", msg)
			}
		}
	}()
}