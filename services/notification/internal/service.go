package notification

import (
	"context"
	"fmt"

	"orderflow/notification/internal/events"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }
func (s *Service) ListByUser(ctx context.Context, userID int64) ([]*Notification, error) {
	return s.repo.ListByUser(ctx, userID)
}
func (s *Service) MarkRead(ctx context.Context, userID, notificationID int64) error {
	return s.repo.MarkRead(ctx, userID, notificationID)
}
func (s *Service) HandleOrderEvent(ctx context.Context, event events.OrderEvent) error {
	if event.EventID == "" ||
		event.EventVersion != events.OrderEventVersion ||
		event.Producer != "order-service" ||
		event.Payload.UserID <= 0 {
		return ErrInvalidEvent
	}
	n := &Notification{UserID: event.Payload.UserID, Type: event.EventType}
	switch event.EventType {
	case events.OrderCreatedEvent:
		n.Title = "Order Created"
		n.Content = "Your order has been created successfully"
	case events.OrderCancelledEvent:
		n.Title = "Order Cancelled"
		n.Content = "Your order has been cancelled"
	default:
		return fmt.Errorf("%w: unsupported type %s", ErrInvalidEvent, event.EventType)
	}
	return s.repo.CreateOnce(ctx, event.EventID, n)
}
