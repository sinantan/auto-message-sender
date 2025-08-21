package dataOperations

import (
	"github.com/sinan/auto-message-sender/internal/config"
	"github.com/sinan/auto-message-sender/pkg/mongodb"
	"github.com/sinan/auto-message-sender/pkg/redisdb"
)

type DataOperations struct {
	mongo  *mongodb.MongoDB
	redis  *redisdb.RedisDB
	config *config.Config
}

func New(mongo *mongodb.MongoDB, redis *redisdb.RedisDB, config *config.Config) *DataOperations {
	return &DataOperations{
		mongo:  mongo,
		redis:  redis,
		config: config,
	}
}
