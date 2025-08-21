package models

import (
	"time"
)

type MessageStatus string

const (
	MessageStatusPending MessageStatus = "pending"
	MessageStatusSent    MessageStatus = "sent"
	MessageStatusFailed  MessageStatus = "failed"
)

type Message struct {
	ID         string        `bson:"_id" json:"id"`
	To         string        `bson:"to" json:"to" validate:"required,e164"`
	Content    string        `bson:"content" json:"content" validate:"required,max=160"`
	Status     MessageStatus `bson:"status" json:"status"`
	CreatedAt  time.Time     `bson:"created_at" json:"created_at"`
	SentAt     *time.Time    `bson:"sent_at,omitempty" json:"sent_at,omitempty"`
	MessageID  *string       `bson:"message_id,omitempty" json:"message_id,omitempty"`
	RetryCount int           `bson:"retry_count" json:"retry_count"`
	Error      *string       `bson:"error,omitempty" json:"error,omitempty"`
}

type WebhookRequest struct {
	To      string `json:"to" validate:"required"`
	Content string `json:"content" validate:"required"`
}

type WebhookResponse struct {
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}

type MessageListResponse struct {
	Messages   []Message `json:"messages"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
}

type CreateMessageRequest struct {
	To      string `json:"to" validate:"required,e164"`
	Content string `json:"content" validate:"required,max=160"`
}

func NewMessage(to, content string) *Message {
	return &Message{
		To:         to,
		Content:    content,
		Status:     MessageStatusPending,
		CreatedAt:  time.Now(),
		RetryCount: 0,
	}
}
