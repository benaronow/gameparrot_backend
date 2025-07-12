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

func handleMessageUpdate(update models.Update) {
	message := updateToMessage(update);
	fmt.Println(update.From);
	fmt.Println(message.To);
	fmt.Println(message.Message);

	var fromUser models.User
	err := mongoClient.UserCollection.FindOne(ctx, map[string]any{"uid": update.From}).Decode(&fromUser)
	if err == mongo.ErrNoDocuments {
		log.Println("Could not find from user");
		return;
	}
	updatedFromFriends := make([]models.Friend, 0, len(fromUser.Friends))
	for _, friend := range fromUser.Friends {
		if friend.UID == update.To {
			friend.Messages = append(friend.Messages, message)
		}
		updatedFromFriends = append(updatedFromFriends, friend)
	}
	fromUpdate := bson.M{
		"$set": bson.M{
			"friends": updatedFromFriends,
		},
	}
	_, err = mongoClient.UserCollection.UpdateOne(ctx, map[string]any{"uid": update.From}, fromUpdate)
	if err != nil {
		log.Println("Could not update from user");
		return;
	}

	var toUser models.User
	err = mongoClient.UserCollection.FindOne(ctx, map[string]any{"uid": update.To}).Decode(&toUser)
	if err == mongo.ErrNoDocuments {
		log.Println("Could not find to user");
		return;
	}
	updatedToFriends := make([]models.Friend, 0, len(toUser.Friends))
	for _, friend := range toUser.Friends {
		if friend.UID == update.From {
			friend.Messages = append(friend.Messages, message)
		}
		updatedToFriends = append(updatedToFriends, friend)
	}
	toUpdate := bson.M{
		"$set": bson.M{
			"friends": updatedToFriends,
		},
	}
	_, err = mongoClient.UserCollection.UpdateOne(ctx, map[string]any{"uid": update.To}, toUpdate)
	if err != nil {
		log.Println("Could not update to user");
		return;
	}
	
	key := fmt.Sprintf("user:%s:online", update.From)
	err = redis.RedisClient.Set(ctx, key, "1", time.Minute).Err()
	messageString, messageErr := json.Marshal(update);
	if err != nil || messageErr != nil {
		log.Println("Redis set online error:", err)
	} else {
		redis.RedisClient.Publish(ctx, "status_channel", "")
		redis.RedisClient.Publish(ctx, "message_channel", messageString)
	}
}

func updateToMessage(update models.Update) models.Message {
	return models.Message{
		Message: update.Message,
		From: update.From,
		To: update.To,
	}
}