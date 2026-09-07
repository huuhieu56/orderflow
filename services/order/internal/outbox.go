package order

import (
	"context"
	"log"
	"time"
)

type OutboxPublisher interface {
	Publish(context.Context, string, string, []byte) error
}
type OutboxWorker struct {
	repo      *Repository
	publisher OutboxPublisher
}

func NewOutboxWorker(repo *Repository, publisher OutboxPublisher) *OutboxWorker {
	return &OutboxWorker{repo: repo, publisher: publisher}
}
func (w *OutboxWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			messages, err := w.repo.PendingOutbox(ctx, 20)
			if err != nil {
				log.Printf("load outbox: %v", err)
				continue
			}
			for _, message := range messages {
				if err := w.publisher.Publish(ctx, message.Topic, message.MessageKey, message.Payload); err != nil {
					log.Printf("publish outbox %s: %v", message.EventID, err)
					_ = w.repo.MarkPublishFailed(ctx, message.ID, 5*time.Second)
					continue
				}
				if err := w.repo.MarkPublished(ctx, message.ID); err != nil {
					log.Printf("mark outbox %s published: %v", message.EventID, err)
				}
			}
		}
	}
}
