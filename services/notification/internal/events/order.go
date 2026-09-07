package events

import "time"

const (
	OrderCreatedEvent   = "order.created"
	OrderCancelledEvent = "order.cancelled"
	OrderEventVersion   = 1
)

type OrderEvent struct {
	EventID      string       `json:"event_id"`
	EventType    string       `json:"event_type"`
	EventVersion int          `json:"event_version"`
	OccurredAt   time.Time    `json:"occurred_at"`
	Producer     string       `json:"producer"`
	Payload      OrderPayload `json:"payload"`
}

type OrderPayload struct {
	OrderID     int64  `json:"order_id"`
	UserID      int64  `json:"user_id"`
	Status      string `json:"status"`
	TotalAmount int64  `json:"total_amount"`
}
