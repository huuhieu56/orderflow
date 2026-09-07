package order

import "time"

const (
	StatusPending   = "pending"
	StatusCancelled = "cancelled"
)

type Order struct {
	ID          int64        `json:"id"`
	UserID      int64        `json:"user_id"`
	Status      string       `json:"status"`
	TotalAmount int64        `json:"total_amount"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Items       []*OrderItem `json:"items"`
}

type OrderItem struct {
	ID          int64  `json:"id"`
	OrderID     int64  `json:"order_id"`
	ProductID   int64  `json:"product_id"`
	ProductName string `json:"product_name"`
	UnitPrice   int64  `json:"unit_price"`
	Quantity    int64  `json:"quantity"`
	Subtotal    int64  `json:"subtotal"`
}

type OutboxMessage struct {
	ID         int64
	EventID    string
	Topic      string
	MessageKey string
	Payload    []byte
}
