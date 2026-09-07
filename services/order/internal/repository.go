package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"orderflow/order/internal/events"
)

type Repository struct {
	db    *sql.DB
	topic string
}

func NewRepository(db *sql.DB, topic string) *Repository { return &Repository{db: db, topic: topic} }

func (r *Repository) Create(ctx context.Context, o *Order, items []*OrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(ctx, `
		INSERT INTO orders (user_id, status, total_amount)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`, o.UserID, o.Status, o.TotalAmount).Scan(
		&o.ID,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		return err
	}
	for _, item := range items {
		item.OrderID = o.ID
		err := tx.QueryRowContext(ctx, `
			INSERT INTO order_items (
				order_id,
				product_id,
				product_name,
				unit_price,
				quantity,
				subtotal
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`,
			o.ID,
			item.ProductID,
			item.ProductName,
			item.UnitPrice,
			item.Quantity,
			item.Subtotal,
		).Scan(&item.ID)
		if err != nil {
			return err
		}
	}
	if err := r.insertOutbox(ctx, tx, newOrderEvent(events.OrderCreatedEvent, o)); err != nil {
		return err
	}
	o.Items = items
	return tx.Commit()
}

func (r *Repository) ListByUser(ctx context.Context, userID int64) ([]*Order, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := make([]*Order, 0)
	for rows.Next() {
		o := &Order{}
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, userID, orderID int64) (*Order, error) {
	o := &Order{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE id = $1 AND user_id = $2
	`, orderID, userID).Scan(
		&o.ID,
		&o.UserID,
		&o.Status,
		&o.TotalAmount,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, order_id, product_id, product_name,
			unit_price, quantity, subtotal
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	o.Items = make([]*OrderItem, 0)
	for rows.Next() {
		item := &OrderItem{}
		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductName,
			&item.UnitPrice,
			&item.Quantity,
			&item.Subtotal,
		); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}
	return o, rows.Err()
}

func (r *Repository) Cancel(ctx context.Context, userID, orderID int64) (*Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	o := &Order{}
	err = tx.QueryRowContext(ctx, `
		UPDATE orders
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status = 'pending'
		RETURNING id, user_id, status, total_amount, created_at, updated_at
	`, orderID, userID).Scan(
		&o.ID,
		&o.UserID,
		&o.Status,
		&o.TotalAmount,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		var exists bool
		err := tx.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM orders WHERE id = $1 AND user_id = $2
			)
		`, orderID, userID).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrOrderNotFound
		}
		return nil, ErrOrderNotCancellable
	}
	if err != nil {
		return nil, err
	}
	if err := r.insertOutbox(ctx, tx, newOrderEvent(events.OrderCancelledEvent, o)); err != nil {
		return nil, err
	}
	return o, tx.Commit()
}

func newOrderEvent(kind string, o *Order) events.OrderEvent {
	return events.OrderEvent{
		EventID:      kind + "-" + strconv.FormatInt(o.ID, 10),
		EventType:    kind,
		EventVersion: events.OrderEventVersion,
		OccurredAt:   time.Now().UTC(),
		Producer:     "order-service",
		Payload: events.OrderPayload{
			OrderID:     o.ID,
			UserID:      o.UserID,
			Status:      o.Status,
			TotalAmount: o.TotalAmount,
		},
	}
}
func (r *Repository) insertOutbox(ctx context.Context, tx *sql.Tx, event events.OrderEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO outbox_events (event_id, topic, message_key, payload)
		VALUES ($1, $2, $3, $4)
	`,
		event.EventID,
		r.topic,
		strconv.FormatInt(event.Payload.OrderID, 10),
		payload,
	)
	return err
}

func (r *Repository) PendingOutbox(ctx context.Context, limit int) ([]OutboxMessage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, event_id, topic, message_key, payload
		FROM outbox_events
		WHERE status = 'pending' AND next_attempt_at <= NOW()
		ORDER BY id
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]OutboxMessage, 0)
	for rows.Next() {
		var m OutboxMessage
		if err := rows.Scan(&m.ID, &m.EventID, &m.Topic, &m.MessageKey, &m.Payload); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}
func (r *Repository) MarkPublished(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE outbox_events SET status='published',published_at=NOW() WHERE id=$1`, id)
	return err
}
func (r *Repository) MarkPublishFailed(ctx context.Context, id int64, retryAfter time.Duration) error {
	_, err := r.db.ExecContext(ctx, `UPDATE outbox_events
		SET attempts=attempts+1,
			status=CASE WHEN attempts+1>=10 THEN 'failed' ELSE 'pending' END,
			next_attempt_at=NOW()+($2*INTERVAL '1 millisecond')
		WHERE id=$1`, id, retryAfter.Milliseconds())
	return err
}
