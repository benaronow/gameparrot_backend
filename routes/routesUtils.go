package routes

import (
	"context"
	"encoding/json"
	"gameparrot_backend/firebase"
	"net/http"

	"firebase.google.com/go/v4/auth"
)

var (
	ctx = context.Background()
)

func getFBToken(w http.ResponseWriter, r *http.Request) *auth.Token {
	var uidReq struct {
		IDToken string `json:"idToken"`
	}

	err := json.NewDecoder(r.Body).Decode(&uidReq)
	if err != nil || uidReq.IDToken == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return nil
	}

	token, err := firebase.AuthClient.VerifyIDToken(ctx, uidReq.IDToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return nil
	}

	return token
}