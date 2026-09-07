package notification

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateOnce(ctx context.Context, eventID string, n *Notification) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `
		INSERT INTO processed_events (event_id)
		VALUES ($1)
		ON CONFLICT DO NOTHING
	`, eventID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return tx.Commit()
	}
	err = tx.QueryRowContext(ctx, `
		INSERT INTO notifications (user_id, type, title, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, n.UserID, n.Type, n.Title, n.Content).Scan(
		&n.ID,
		&n.CreatedAt,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) ListByUser(ctx context.Context, userID int64) ([]*Notification, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, type, title, content, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	notifications := make([]*Notification, 0)
	for rows.Next() {
		n := &Notification{}
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Content, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (r *Repository) MarkRead(ctx context.Context, userID, notificationID int64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE notifications
		SET is_read = TRUE
		WHERE id = $1 AND user_id = $2
	`, notificationID, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotificationNotFound
	}
	return nil
}
