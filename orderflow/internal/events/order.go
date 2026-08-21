package events

import "time"

const (
	OrderCreatedEvent   = "order.created"
	OrderCancelledEvent = "order.cancelled"
)

type OrderEvent struct {
	EventID   string    `json:"event_id"`
	Type      string    `json:"type"`
	OrderID   int64     `json:"order_id"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
