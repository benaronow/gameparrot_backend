package routes

import (
	"context"
	"encoding/json"
	"gameparrot_backend/firebase"
	"net/http"
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
    w.Write([]byte(uid));
}