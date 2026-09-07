package order

import (
	"context"
	"math"
)

type Service struct {
	repo          *Repository
	productFinder ProductFinder
}

func NewService(repo *Repository, finder ProductFinder) *Service {
	return &Service{repo: repo, productFinder: finder}
}

type CreateOrderItem struct {
	ProductID int64
	Quantity  int64
}

func (s *Service) CreateOrder(ctx context.Context, userID int64, input []CreateOrderItem) (*Order, error) {
	if userID <= 0 || len(input) == 0 {
		return nil, ErrInvalidOrder
	}
	quantities := make(map[int64]int64, len(input))
	productIDs := make([]int64, 0, len(input))
	for _, item := range input {
		if item.ProductID <= 0 || item.Quantity <= 0 {
			return nil, ErrInvalidOrder
		}
		if _, exists := quantities[item.ProductID]; !exists {
			productIDs = append(productIDs, item.ProductID)
		}
		quantities[item.ProductID] += item.Quantity
	}
	o := &Order{UserID: userID, Status: StatusPending}
	items := make([]*OrderItem, 0, len(quantities))
	for _, productID := range productIDs {
		quantity := quantities[productID]
		p, err := s.productFinder.GetProduct(ctx, productID)
		if err != nil {
			return nil, err
		}
		if p.Status != "active" {
			return nil, ErrProductUnavailable
		}
		if p.ID != productID || p.Price <= 0 || p.Stock < 0 {
			return nil, ErrProductUnavailable
		}
		if p.Stock < quantity {
			return nil, ErrInsufficientStock
		}
		if quantity > math.MaxInt64/p.Price {
			return nil, ErrInvalidOrder
		}
		subtotal := p.Price * quantity
		if o.TotalAmount > math.MaxInt64-subtotal {
			return nil, ErrInvalidOrder
		}
		items = append(items, &OrderItem{
			ProductID:   p.ID,
			ProductName: p.Name,
			UnitPrice:   p.Price,
			Quantity:    quantity,
			Subtotal:    subtotal,
		})
		o.TotalAmount += subtotal
	}
	if err := s.repo.Create(ctx, o, items); err != nil {
		return nil, err
	}
	return o, nil
}
func (s *Service) GetOrder(ctx context.Context, userID, orderID int64) (*Order, error) {
	return s.repo.GetByID(ctx, userID, orderID)
}
func (s *Service) ListOrdersByUser(ctx context.Context, userID int64) ([]*Order, error) {
	return s.repo.ListByUser(ctx, userID)
}
func (s *Service) CancelOrder(ctx context.Context, userID, orderID int64) (*Order, error) {
	return s.repo.Cancel(ctx, userID, orderID)
}
