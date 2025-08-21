package dataOperations

import (
	"github.com/sinan/auto-message-sender/internal/models"
	"github.com/sinan/auto-message-sender/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const SchedulerLogsCollection = "scheduler_logs"

func (do *DataOperations) CreateSchedulerStartLog() (string, error) {
	log := models.NewSchedulerStartLog()
	err := mongodb.InsertOne(do.mongo, SchedulerLogsCollection, log)
	return log.ID, err
}

func (do *DataOperations) CreateSchedulerStopLog(startID string) error {
	log := models.NewSchedulerStopLog(startID)
	return mongodb.InsertOne(do.mongo, SchedulerLogsCollection, log)
}

func (do *DataOperations) IsSchedulerActive() (bool, error) {
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(1)
	results, err := mongodb.Query[models.SchedulerLog](do.mongo, SchedulerLogsCollection, bson.M{}, opts)

	if err != nil {
		return false, err
	}

	if len(results) == 0 {
		return false, nil
	}

	return results[0].Action == models.SchedulerActionStart, nil
}

func (do *DataOperations) StartScheduler() (string, error) {
	return do.CreateSchedulerStartLog()
}

func (do *DataOperations) StopScheduler(startID string) error {
	return do.CreateSchedulerStopLog(startID)
}
