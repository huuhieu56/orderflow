package notification

import (
	"orderflow/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListByUser(userID int64) ([]*models.Notification, error) {
	return s.repo.ListByUser(userID)
}

func (s *Service) MarkRead(userID, notificationID int64) error {
	return s.repo.MarkRead(userID, notificationID)
}

func (s *Service) CreateOrderCreated(order *models.Order) error {
	return s.repo.Create(&models.Notification{
		UserID: order.UserID,
		Type: "order.created",
		Title: "Order Created",
		Content: "Your order has been created successfully",
	})
}

func (s* Service) CreateOrderCancelled(order *models.Order) error {
	return s.repo.Create(&models.Notification{
		UserID: order.UserID,
		Type: "order.cancelled",
		Title: "Order Cancelled",
		Content: "Your order has been cancelled",
	})
}

