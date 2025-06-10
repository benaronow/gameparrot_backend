package firebase

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var AuthClient *auth.Client

func InitFirebase() {
	ctx := context.Background()
	opt := option.WithCredentialsJSON([]byte(os.Getenv("FIREBASE_SECRET")))

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
	}

	AuthClient, err = app.Auth(ctx)
	if err != nil {
		log.Fatalf("Error getting Firebase Auth client: %v", err)
	}
}