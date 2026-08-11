package order

import (
	"errors"
	"orderflow/internal/models"
	"orderflow/internal/product"
)

type Service struct {
	repo        *Repository
	productRepo *product.Repository
}

func NewService(repo *Repository, productRepo *product.Repository) *Service {
	return &Service{repo: repo, productRepo: productRepo}
}

type CreateOrderItem struct {
	ProductID int64
	Quantity  int64
}

func (s *Service) CreateOrder(userID int64, items []CreateOrderItem) (*models.Order, error) {
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	order := &models.Order{
		UserID: userID,
		Status: "pending",
	}

	var orderItems []*models.OrderItem
	var total float64

	for _, it := range items {
		p, err := s.productRepo.GetByID(it.ProductID)
		if err != nil {
			return nil, err
		}
		if p.Status != "active" {
			return nil, errors.New("product is not available")
		}

		subtotal := p.Price * float64(it.Quantity)
		orderItems = append(orderItems, &models.OrderItem{
			ProductID:   p.ID,
			ProductName: p.Name,
			UnitPrice:   p.Price,
			Quantity:    it.Quantity,
			Subtotal:    subtotal,
		})
		total += subtotal
	}

	order.TotalAmount = total

	if err := s.repo.Create(order, orderItems); err != nil {
		return nil, err
	}
	return order, nil
}
