package models

import "time"

type Order struct {
	ID          int64
	UserID      int64
	Status      string
	TotalAmount float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Items 		[]*OrderItem
}

type OrderItem struct {
	ID          int64
	OrderID     int64
	ProductID   int64
	ProductName string
	UnitPrice   float64
	Quantity    int64
	Subtotal    float64
}
