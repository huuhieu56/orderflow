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


func (r *Repository) GetByID(userID, orderID int64) (*models.Order, error) {
	query := 
	`
		SELECT id, user_id, status, total_amount, created_at, updated_at 
		FROM orders
		WHERE id = $1 AND user_id = $2 
	`

	order := &models.Order{}

	err := r.db.QueryRow(
		query,
		orderID,
		userID,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(`
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

	order.Items = make([]*models.OrderItem, 0)

	for rows.Next() {
		item := &models.OrderItem{}

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

		order.Items = append(order.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err 
	}

	return order, nil 
}

func (r *Repository) Cancel(userID, orderID int64) (*models.Order, error) {
	order := &models.Order{} 

	err := r.db.QueryRow(`
		UPDATE orders
		SET status = 'cancelled', updated_at = NOW() 
		WHERE id = $1 
			AND user_id = $2 
			AND status = 'pending'
		RETURNING id, user_id, status, total_amount, created_at, updated_at	
	`, orderID, userID).Scan(
		&order.ID, 
		&order.UserID, 
		&order.Status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		return nil, err 
	}

	return order, nil 
}
