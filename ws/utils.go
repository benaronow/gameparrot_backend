package ws

import (
	"context"
	"gameparrot_backend/models"
)

var (
	ctx = context.Background()
)

func updateToMessage(update models.Update) models.Message {
	return models.Message{
		Message: update.Message,
		From: update.From,
		To: update.To,
	}
}

func updateToFriendRequest(update models.Update) models.FriendRequest {
	return models.FriendRequest{
		From: update.From,
		To: update.To,
	}
}