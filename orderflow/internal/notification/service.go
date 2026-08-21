package notification

import (
	"fmt"

	"orderflow/internal/events"
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

func (s *Service) HandleOrderEvent(event events.OrderEvent) error {
	notification := &models.Notification{
		UserID: event.UserID,
		Type:   event.Type,
	}

	switch event.Type {
	case events.OrderCreatedEvent:
		notification.Title = "Order Created"
		notification.Content = "Your order has been created successfully"
	case events.OrderCancelledEvent:
		notification.Title = "Order Cancelled"
		notification.Content = "Your order has been cancelled"
	default:
		return fmt.Errorf("unsupported order event type: %s", event.Type)
	}

	return s.repo.Create(notification)
}
