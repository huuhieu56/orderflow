package product

import "time"

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	Stock       int64     `json:"stock"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateProduct struct {
	Name        *string
	Description *string
	Price       *int64
	Stock       *int64
	Status      *string
}
