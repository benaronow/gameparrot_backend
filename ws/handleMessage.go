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

	var fromUser models.User
	err := mongoClient.UserCollection.FindOne(ctx, map[string]any{"uid": update.From}).Decode(&fromUser)
	if err == mongo.ErrNoDocuments {
		log.Println("Could not find from user");
		return;
	}
	updatedFromInteractions := make([]models.Interaction, 0, len(fromUser.Interactions))
	for _, interaction := range fromUser.Interactions {
		if interaction.UID == update.To {
			interaction.Messages = append(interaction.Messages, message)
		}
		updatedFromInteractions = append(updatedFromInteractions, interaction)
	}
	fromUpdate := bson.M{
		"$set": bson.M{
			"interactions": updatedFromInteractions,
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
	updatedToInteractions := make([]models.Interaction, 0, len(toUser.Interactions))
	for _, interaction := range toUser.Interactions {
		if interaction.UID == update.From {
			interaction.Messages = append(interaction.Messages, message)
		}
		updatedToInteractions = append(updatedToInteractions, interaction)
	}
	toUpdate := bson.M{
		"$set": bson.M{
			"interactions": updatedToInteractions,
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