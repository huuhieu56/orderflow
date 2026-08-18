package order

import (
	"database/sql"
	"errors"
	"orderflow/internal/models"
	"orderflow/internal/notification"
	"orderflow/internal/product"
)

type Service struct {
	repo            *Repository
	productRepo     *product.Repository
	notificationSvc *notification.Service
}

var ErrOrderNotCancellable = errors.New("order cannot be cancelled")

func NewService(repo *Repository, productRepo *product.Repository, notificationSvc *notification.Service) *Service {
	return &Service{repo: repo,
		productRepo:     productRepo,
		notificationSvc: notificationSvc,
	}
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
		if it.Quantity <= 0 {
			return nil, errors.New("quantity must be greater than zero")
		}

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

	if err := s.notificationSvc.CreateOrderCreated(order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *Service) GetOrder(userID, orderID int64) (*models.Order, error) {
	return s.repo.GetByID(userID, orderID)
}

func (s *Service) ListOrdersByUser(userID int64) ([]*models.Order, error) {
	return s.repo.GetByUserID(userID)
}

func (s *Service) CancelOrder(userID, orderID int64) (*models.Order, error) {
	order, err := s.repo.GetByID(userID, orderID)
	if err != nil {
		return nil, err
	}

	if order.Status != "pending" {
		return nil, ErrOrderNotCancellable
	}

	cancelled, err := s.repo.Cancel(userID, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOrderNotCancellable
	}

	if err != nil {
		return nil, err
	}

	if err := s.notificationSvc.CreateOrderCancelled(cancelled); err != nil {
		return nil, err
	}

	return cancelled, nil
}
