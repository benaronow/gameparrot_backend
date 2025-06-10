package main

import (
	"fmt"
	"gameparrot_backend/firebase"
	"gameparrot_backend/middleware"
	"gameparrot_backend/mongo"
	"gameparrot_backend/redis"
	"gameparrot_backend/routes"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
	
	redis.InitRedis()
    firebase.InitFirebase()
    mongo.InitMongo()

    go redis.StartRedisSubscriber()

    http.HandleFunc("/ws", middleware.WithCORS(routes.WebSocketHandler))
    http.HandleFunc("/auth", middleware.WithCORS(routes.AuthHandler))
    http.HandleFunc("/register", middleware.WithCORS(routes.RegisterHandler))
    fmt.Println("WebSocket server running at :8080/ws")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
