package notification

import (
	"database/sql"
	"orderflow/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(n *models.Notification) error {
	return r.db.QueryRow(`
		INSERT INTO notifications (user_id, type, title, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at 
	`, n.UserID, n.Type, n.Title, n.Content).Scan(
		&n.ID,
		&n.CreatedAt,
	)
}
func (r *Repository) ListByUser(userID int64) ([]*models.Notification, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, type, title, content, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]*models.Notification, 0)

	for rows.Next() {
		n := &models.Notification{}

		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&n.Title,
			&n.Content,
			&n.IsRead,
			&n.CreatedAt,
		); err != nil {
			return nil, err
		}

		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}

func (r *Repository) MarkRead(userID, notificationID int64) error {
	result, err := r.db.Exec(`
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
		return sql.ErrNoRows
	}

	return nil
}
