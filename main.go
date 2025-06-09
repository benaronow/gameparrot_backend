package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/option"
)

var (
    ctx = context.Background()
    redisClient *redis.Client
    upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
    }
    clients = make(map[*websocket.Conn]bool)
    clientsLock sync.Mutex
    authClient *auth.Client
	userCollection *mongo.Collection
)

func main() {
	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
	
	redisURI := os.Getenv("REDIS_URI")
	opt, _ := redis.ParseURL(redisURI)
    redisClient = redis.NewClient(opt)

    initFirebase()
    initMongo()

    go startRedisSubscriber()

    http.HandleFunc("/ws", withCORS(wsHandler))
    http.HandleFunc("/auth", withCORS(authHandler))
    http.HandleFunc("/register", withCORS(registerHandler))
    fmt.Println("WebSocket server running at :8080/ws")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the actual handler
		h(w, r)
	}
}

func initFirebase() {
	ctx := context.Background()
    firebaseSecret := os.Getenv("FIREBASE_SECRET")
	opt := option.WithCredentialsJSON([]byte(firebaseSecret))

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v\n", err)
	}

	authClient, err = app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}
}

func initMongo() {
    mongoURI := os.Getenv("MONGO_URI")
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("MongoDB connection error: %v\n", err)
	}
	userCollection = client.Database("gameparrot").Collection("User")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDToken string `json:"idToken"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.IDToken == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	token, err := authClient.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	uid := token.UID
	email := token.Claims["email"]
    println(uid, email)

	// Check if user already exists
	var existing map[string]interface{}
	err = userCollection.FindOne(ctx, map[string]interface{}{"uid": uid}).Decode(&existing)
    log.Printf("Object: %+v\n", err)
	if err == mongo.ErrNoDocuments {
		// New user — insert
		newUser := map[string]interface{}{
			"uid":   uid,
			"email": email,
			"createdAt": token.IssuedAt,
		}
		_, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			http.Error(w, "Failed to insert user", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "User registered successfully")
		return
	}

	// Already exists
	fmt.Fprintln(w, "User already exists")
}

func authHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        IDToken string `json:"idToken"`
    }

    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    ctx := context.Background()
    token, err := authClient.VerifyIDToken(ctx, req.IDToken)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    uid := token.UID
    w.Write([]byte(uid));
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("WebSocket upgrade error:", err)
        return
    }
    defer conn.Close()

    // Register client
    clientsLock.Lock()
    clients[conn] = true
    clientsLock.Unlock()
    log.Println("Client connected")

    // Handle incoming messages
    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            log.Println("WebSocket read error:", err)
            break
        }

        // Publish to Redis
        err = redisClient.Publish(ctx, "game_channel", msg).Err()
        if err != nil {
            log.Println("Redis publish error:", err)
        }
    }

    // Cleanup on disconnect
    clientsLock.Lock()
    delete(clients, conn)
    clientsLock.Unlock()
    log.Println("Client disconnected")
}

func startRedisSubscriber() {
    sub := redisClient.Subscribe(ctx, "game_channel")
    ch := sub.Channel()
	

    for msg := range ch {
        broadcast([]byte(msg.Payload))
    }
}

func broadcast(message []byte) {
    clientsLock.Lock()
    defer clientsLock.Unlock()

    for conn := range clients {
        err := conn.WriteMessage(websocket.TextMessage, message)
        if err != nil {
            log.Println("Broadcast error:", err)
            conn.Close()
            delete(clients, conn)
        }
    }
}
