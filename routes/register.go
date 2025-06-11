package routes

import (
	"fmt"
	mongoClient "gameparrot_backend/mongo"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var token = getFBToken(w, r)

	var existing map[string]any
	err := mongoClient.UserCollection.FindOne(ctx, map[string]any{"uid": token.UID}).Decode(&existing)
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