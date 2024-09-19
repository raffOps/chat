package chat

import (
	"time"

	"github.com/google/uuid"
)

type Topic struct {
	Name   string
	Offset int32
}

type Message struct {
	Id          uuid.UUID `json:"id" validate:"required"`
	ChatId      uuid.UUID `json:"chat_id" validate:"required"`
	Offset      int64     `json:"offset" validate:"required"`
	DeliveredAt time.Time `json:"delivered_at,omitempty" validate:"required"`
	CreatedAt   time.Time `json:"created_at" validate:"required"`
	From        int64     `json:"from" validate:"required"`
	To          int64     `json:"to" validate:"required"`
	Content     string    `json:"content" validate:"required"`
}

var (
	MaxDeliveryDelay    = "100s"
	MessageWriteTimeout = "1000s"
	MessageReadTimeout  = "10s"
)
