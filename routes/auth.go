package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"gameparrot_backend/firebase"
	"gameparrot_backend/redis"
	"log"
	"net/http"
	"time"
)

func AuthHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        IDToken string `json:"idToken"`
    }

    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    ctx := context.Background()
    token, err := firebase.AuthClient.VerifyIDToken(ctx, req.IDToken)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    uid := token.UID
    key := fmt.Sprintf("user:%s:online", uid)
	redisErr := redis.RedisClient.Set(context.Background(), key, 1, time.Minute*10).Err()
    if redisErr != nil {
        log.Println("Failed to set user online status in Redis:", err)
    }
    
    w.Write([]byte(uid));
}