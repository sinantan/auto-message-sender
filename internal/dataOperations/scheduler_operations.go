package dataOperations

import (
	"github.com/sinan/auto-message-sender/internal/models"
	"github.com/sinan/auto-message-sender/pkg/mongodb"
)

const SchedulerCollection = "scheduler_status"

func (do *DataOperations) GetSchedulerStatus() (*models.SchedulerStatus, error) {
	status, err := mongodb.GetOneById[models.SchedulerStatus](do.mongo, SchedulerCollection, models.SchedulerStatusID)
	if err != nil {
		return nil, err
	}

	if status == nil {
		defaultStatus := models.NewSchedulerStatus(false)
		err = mongodb.UpsertOne(do.mongo, SchedulerCollection, defaultStatus.ID, defaultStatus)
		if err != nil {
			return nil, err
		}
		return defaultStatus, nil
	}

	return status, nil
}

func (do *DataOperations) UpdateSchedulerStatus(isActive bool) error {
	status := models.NewSchedulerStatus(isActive)
	return mongodb.UpsertOne(do.mongo, SchedulerCollection, status.ID, status)
}
