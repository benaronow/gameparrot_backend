package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"gameparrot_backend/firebase"
	mongoClient "gameparrot_backend/mongo"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDToken string `json:"idToken"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.IDToken == "" {
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

	var existing map[string]any
	err = mongoClient.UserCollection.FindOne(ctx, map[string]any{"uid": uid}).Decode(&existing)
	if err == mongo.ErrNoDocuments {
		newUser := map[string]any{
			"uid":   token.UID,
			"email": token.Claims["email"],
			"createdAt": token.IssuedAt,
		}
		_, err := mongoClient.UserCollection.InsertOne(ctx, newUser)
		if err != nil {
			http.Error(w, "Failed to insert user", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "User registered successfully")
		return
	}

	fmt.Fprintln(w, "User already exists")
}