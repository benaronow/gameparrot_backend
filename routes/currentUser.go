package routes

import (
	"encoding/json"
	"gameparrot_backend/models"
	mongoClient "gameparrot_backend/mongo"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

func CurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		http.Error(w, "Missing uid", http.StatusBadRequest)
		return
	}

	var user models.User
	err := mongoClient.UserCollection.FindOne(ctx, map[string]any{"uid": uid}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		http.Error(w, "User not found", http.StatusNotFound)
		return;
	}

	data, err := json.Marshal(user)
    if err != nil {
		http.Error(w, "Error parsing response", http.StatusBadRequest)
        log.Println("JSON marshal error:", err)
		return;
    }

    w.Write([]byte(data));
}