package main

import (
	"fmt"
	"gameparrot_backend/firebase"
	"gameparrot_backend/middleware"
	"gameparrot_backend/mongo"
	"gameparrot_backend/redis"
	"gameparrot_backend/routes"
	"gameparrot_backend/ws"
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
	go redis.StartStatusRebroadcast()

    http.HandleFunc("/auth", middleware.WithCORS(routes.AuthHandler))
    http.HandleFunc("/register", middleware.WithCORS(routes.RegisterHandler))
	http.HandleFunc("/currentUser", middleware.WithCORS(routes.CurrentUserHandler))
	http.HandleFunc("/ws", middleware.WithCORS(ws.HandleWSConnection))
	
    fmt.Println("WebSocket server running at :8080/ws")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
