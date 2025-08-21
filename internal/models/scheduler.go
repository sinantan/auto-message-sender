package models

import "time"

type SchedulerStatus struct {
	ID        string     `bson:"_id" json:"id"`
	IsActive  bool       `bson:"is_active" json:"is_active"`
	StartedAt *time.Time `bson:"started_at,omitempty" json:"started_at,omitempty"`
	StoppedAt *time.Time `bson:"stopped_at,omitempty" json:"stopped_at,omitempty"`
	UpdatedAt time.Time  `bson:"updated_at" json:"updated_at"`
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

const SchedulerStatusID = "scheduler_status"

func NewSchedulerStatus(isActive bool) *SchedulerStatus {
	now := time.Now()
	status := &SchedulerStatus{
		ID:        SchedulerStatusID,
		IsActive:  isActive,
		UpdatedAt: now,
	}

	if isActive {
		status.StartedAt = &now
	} else {
		status.StoppedAt = &now
	}

	return status
}
