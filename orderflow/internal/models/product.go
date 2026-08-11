package models

import "time"

type Product struct {
	ID          int64
	Name        string
	Description string
	Price       float64
	Stock       int64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
