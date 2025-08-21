package dataOperations

import (
	"context"
	"fmt"
	"time"

	"github.com/sinan/auto-message-sender/internal/models"
)

func (do *DataOperations) CacheMessage(messageID string, sentAt time.Time) error {
	key := fmt.Sprintf("message_sent:%s", messageID)
	cachedMsg := models.CachedMessage{
		MessageID: messageID,
		SentAt:    sentAt,
	}

	return do.redis.SetJSON(context.Background(), key, cachedMsg, 24*time.Hour)
}

func (do *DataOperations) GetCachedMessage(messageID string) (*models.CachedMessage, error) {
	key := fmt.Sprintf("message_sent:%s", messageID)

	var cachedMsg models.CachedMessage
	err := do.redis.GetJSON(context.Background(), key, &cachedMsg)
	if err != nil {
		return nil, err
	}

	return &cachedMsg, nil
}

func (do *DataOperations) IsCachedMessage(messageID string) (bool, error) {
	key := fmt.Sprintf("message_sent:%s", messageID)
	return do.redis.Exists(context.Background(), key)
}
