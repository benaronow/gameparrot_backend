package routes

import (
	"fmt"
	"gameparrot_backend/redis"
	"log"
	"net/http"
	"time"
)

func AuthHandler(w http.ResponseWriter, r *http.Request) {
    var token = getFBToken(w, r)

    key := fmt.Sprintf("user:%s:online", token.UID)
	redisErr := redis.RedisClient.Set(ctx, key, 1, time.Minute*10).Err()
    if redisErr != nil {
        log.Println("Failed to set user online status in Redis:", redisErr)
    }

    w.Write([]byte(token.UID));
}