package product

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

func (r *Repository) Create(p *models.Product) error {
	query := `
              INSERT INTO products (name, description, price, stock, status)
              VALUES ($1, $2, $3, $4, $5)
              RETURNING id, created_at, updated_at
      `
	err := r.db.QueryRow(
		query, p.Name, p.Description, p.Price, p.Stock, p.Status,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	return err
}

func (r *Repository) GetByID(id int64) (*models.Product, error) {
	query := `
              SELECT id, name, description, price, stock, status, created_at, updated_at
              FROM products WHERE id = $1
      `
	p := &models.Product{}
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock,
		&p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	return p, err
}

func (r *Repository) List(onlyActive bool) ([]*models.Product, error) {
	query := `SELECT id, name, description, price, stock, status, created_at, updated_at FROM products`
	if onlyActive {
		query += ` WHERE status = 'active'`
	}
	query += ` ORDER BY id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		p := &models.Product{}
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock,
			&p.Status, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}
