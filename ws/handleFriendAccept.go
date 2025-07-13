package ws

import (
	"encoding/json"
	"fmt"
	"gameparrot_backend/models"
	mongoClient "gameparrot_backend/mongo"
	"gameparrot_backend/redis"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func handleFriendAcceptUpdate(update models.Update) {
	friendRequest := updateToFriendRequest(update)

	var fromUser models.User
	err := mongoClient.UserCollection.FindOne(ctx, map[string]any{"uid": update.From}).Decode(&fromUser)
	if err == mongo.ErrNoDocuments {
		log.Println("Could not find from user")
		return
	}
	for i, req := range fromUser.FriendRequests {
		if req.From == friendRequest.From && req.To == friendRequest.To {
			fromUser.FriendRequests = append(fromUser.FriendRequests[:i], fromUser.FriendRequests[i+1:]...)
			break
		}
	}
	fromUser.Interactions = append(fromUser.Interactions, models.Interaction{
		UID:     update.To,
		Messages: []models.Message{},
	})
	fromUpdate := bson.M{
		"$set": bson.M{
			"interactions": fromUser.Interactions,
			"friend_requests": fromUser.FriendRequests,
		},
	}
	_, err = mongoClient.UserCollection.UpdateOne(ctx, map[string]any{"uid": update.From}, fromUpdate)
	if err != nil {
		log.Println("Could not update from user")
		return
	}

	var toUser models.User
	err = mongoClient.UserCollection.FindOne(ctx, map[string]any{"uid": update.To}).Decode(&toUser)
	if err == mongo.ErrNoDocuments {
		log.Println("Could not find to user")
		return
	}
	for i, req := range toUser.FriendRequests {
		if req.From == friendRequest.From && req.To == friendRequest.To {
			toUser.FriendRequests = append(toUser.FriendRequests[:i], toUser.FriendRequests[i+1:]...)
			break
		}
	}
	toUser.Interactions = append(toUser.Interactions, models.Interaction{
		UID:     update.From,
		Messages: []models.Message{},
	})
	toUpdate := bson.M{
		"$set": bson.M{
			"interactions": toUser.Interactions,
			"friend_requests": toUser.FriendRequests,
		},
	}
	_, err = mongoClient.UserCollection.UpdateOne(ctx, map[string]any{"uid": update.To}, toUpdate)
	if err != nil {
		log.Println("Could not update to user")
		return
	}

	key := fmt.Sprintf("user:%s:online", update.To)
	err = redis.RedisClient.Set(ctx, key, "1", time.Minute).Err()
	friendAcceptString, friendAcceptErr := json.Marshal(update)
	if err != nil || friendAcceptErr != nil {
		log.Println("Redis set online error:", err)
	} else {
		redis.RedisClient.Publish(ctx, "status_channel", "")
		redis.RedisClient.Publish(ctx, "friend_accept_channel", friendAcceptString)
	}
}