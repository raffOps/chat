package chat

import "time"

type Topic struct {
	Name   string
	Offset int32
}

type Message struct {
	Id          string    `json:"id" validate:"required"`
	Topic       string    `json:"topic" validate:"required"`
	Offset      int64     `json:"offset" validate:"required"`
	DeliveredAt time.Time `json:"delivered_at,omitempty" validate:"required"`
	CreatedAt   time.Time `json:"created_at" validate:"required"`
	From        int64     `json:"from" validate:"required"`
	To          int64     `json:"to" validate:"required"`
	Content     string    `json:"content" validate:"required"`
}

var (
	MessageWriteTimeout = "10s"
	MessageReadTimeout  = "10s"
)
