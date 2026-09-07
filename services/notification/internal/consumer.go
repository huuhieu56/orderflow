package notification

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"orderflow/notification/internal/events"
	"orderflow/notification/internal/messaging"
)

type EventConsumer struct {
	service  *Service
	consumer *messaging.Consumer
}

func NewEventConsumer(
	service *Service,
	consumer *messaging.Consumer,
) *EventConsumer {
	return &EventConsumer{
		service:  service,
		consumer: consumer,
	}
}

func (c *EventConsumer) Run(ctx context.Context) error {
	for {
		message, err := c.consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("kafka fetch failed, retrying: %v", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
			}
			continue
		}

		var event events.OrderEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("invalid order event: %v", err)

			if err := c.consumer.CommitMessage(ctx, message); err != nil {
				log.Printf("kafka commit failed: %v", err)
			}
			continue
		}

		if err := c.service.HandleOrderEvent(ctx, event); err != nil {
			log.Printf(
				"failed to handle event %s: %v",
				event.EventID,
				err,
			)
			if !errors.Is(err, ErrInvalidEvent) {
				continue
			}
		}

		if err := c.consumer.CommitMessage(ctx, message); err != nil {
			log.Printf("kafka commit failed: %v", err)
			continue
		}

	}
}
