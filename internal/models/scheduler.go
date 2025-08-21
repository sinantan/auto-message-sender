package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type SchedulerAction string

const (
	SchedulerActionStart SchedulerAction = "start"
	SchedulerActionStop  SchedulerAction = "stop"
)

type SchedulerLog struct {
	ID        string          `bson:"_id" json:"id"`
	Action    SchedulerAction `bson:"action" json:"action"`
	StartID   *string         `bson:"start_id,omitempty" json:"start_id,omitempty"`
	Timestamp time.Time       `bson:"timestamp" json:"timestamp"`
	CreatedAt time.Time       `bson:"created_at" json:"created_at"`
}

type SchedulerResponse struct {
	IsActive  bool       `json:"is_active"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
	Message   string     `json:"message"`
}

type CachedMessage struct {
	MessageID string    `json:"message_id"`
	SentAt    time.Time `json:"sent_at"`
}

func NewSchedulerStartLog() *SchedulerLog {
	now := time.Now()
	return &SchedulerLog{
		ID:        primitive.NewObjectID().Hex(),
		Action:    SchedulerActionStart,
		StartID:   nil,
		Timestamp: now,
		CreatedAt: now,
	}
}

func NewSchedulerStopLog(startID string) *SchedulerLog {
	now := time.Now()
	return &SchedulerLog{
		ID:        primitive.NewObjectID().Hex(),
		Action:    SchedulerActionStop,
		StartID:   &startID,
		Timestamp: now,
		CreatedAt: now,
	}
}
