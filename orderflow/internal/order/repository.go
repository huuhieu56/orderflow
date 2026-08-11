package order

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

func (r *Repository) Create(o *models.Order, items []*models.OrderItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// Order trước, lấy id
	err = tx.QueryRow(
		`INSERT INTO orders (user_id, status, total_amount)
               VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`,
		o.UserID, o.Status, o.TotalAmount,
	).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Items, mỗi dòng lấy order_id vừa tạo
	for _, item := range items {
		err = tx.QueryRow(
			`INSERT INTO order_items (order_id, product_id, product_name, unit_price, quantity, subtotal)
                       VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			o.ID, item.ProductID, item.ProductName,
			item.UnitPrice, item.Quantity, item.Subtotal,
		).Scan(&item.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) GetByUserID(userID int64) ([]*models.Order, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, status, total_amount, created_at, updated_at
               FROM orders WHERE user_id = $1 ORDER BY id`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		o := &models.Order{}
		if err := rows.Scan(
			&o.ID, &o.UserID, &o.Status, &o.TotalAmount,
			&o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
