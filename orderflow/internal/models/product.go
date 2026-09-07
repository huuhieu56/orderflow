package models

import "time"

type Product struct {
	ID          int64		`json:"id"`
	Name        string		`json:"name"`
	Description string		`json:"description"`
	Price       float64		`json:"price"`
	Stock       int64		`json:"stock"`
	Status      string		`json:"status"`
	CreatedAt   time.Time	`json:"created_at"`
	UpdatedAt   time.Time	`json:"updated_at"`
}
