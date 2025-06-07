package main

import (
	"fmt"
	"log"
	"net/http"
	"go-server/handlers"
)

func main() {
	http.HandleFunc("/ws", handlers.WebSocketHandler)
	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}