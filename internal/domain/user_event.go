package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type UserEvent struct {
	ID        uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID       `gorm:"type:uuid;not null;index"`
	EventType string          `gorm:"not null"`
	Payload   json.RawMessage `gorm:"type:jsonb"`
	CreatedAt time.Time
}

func (UserEvent) TableName() string {
	return "user_events"
}

type EventPayload struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

