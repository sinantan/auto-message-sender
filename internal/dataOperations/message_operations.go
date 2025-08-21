package dataOperations

import (
	"context"
	"github.com/sinan/auto-message-sender/internal/models"
	"github.com/sinan/auto-message-sender/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const MessagesCollection = "messages"

func (do *DataOperations) CreateMessage(message *models.Message) error {
	return mongodb.InsertOne(do.mongo, MessagesCollection, message)
}

func (do *DataOperations) GetPendingMessages(limit int) ([]models.Message, error) {
	filter := bson.M{
		"status":      models.MessageStatusPending,
		"retry_count": bson.M{"$lt": do.config.App.MaxRetryCount},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetLimit(int64(limit))

	return mongodb.Query[models.Message](do.mongo, MessagesCollection, filter, opts)
}

func (do *DataOperations) GetSentMessages(page, perPage int) ([]models.Message, int64, error) {
	filter := bson.M{"status": models.MessageStatusSent}

	total, err := mongodb.Count(do.mongo, MessagesCollection, filter, nil)
	if err != nil {
		return nil, 0, err
	}

	skip := (page - 1) * perPage
	opts := options.Find().
		SetSort(bson.D{{Key: "sent_at", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(perPage))

	messages, err := mongodb.Query[models.Message](do.mongo, MessagesCollection, filter, opts)
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

func (do *DataOperations) UpdateMessageStatus(messageID string, status models.MessageStatus, webhookMessageID *string, errorMsg *string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if status == models.MessageStatusSent && webhookMessageID != nil {
		update["$set"].(bson.M)["sent_at"] = time.Now()
		update["$set"].(bson.M)["message_id"] = *webhookMessageID
	}

	if status == models.MessageStatusFailed {
		update["$inc"] = bson.M{"retry_count": 1}
		if errorMsg != nil {
			update["$set"].(bson.M)["error"] = *errorMsg
		}
	}

	filter := bson.M{"_id": messageID}
	_, err := do.mongo.GetClient()
	if err != nil {
		return err
	}

	client, err := do.mongo.GetClient()
	if err != nil {
		return err
	}

	collection := client.Database(do.mongo.DBName).Collection(MessagesCollection)
	_, err = collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (do *DataOperations) GetMessageByID(messageID string) (*models.Message, error) {
	return mongodb.GetOneById[models.Message](do.mongo, MessagesCollection, messageID)
}
