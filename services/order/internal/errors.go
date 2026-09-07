package order

import "errors"

var (
	ErrInvalidOrder        = errors.New("invalid order")
	ErrOrderNotFound       = errors.New("order not found")
	ErrOrderNotCancellable = errors.New("order cannot be cancelled")
	ErrProductUnavailable  = errors.New("product is unavailable")
	ErrInsufficientStock   = errors.New("insufficient stock")
)
