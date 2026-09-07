package product

import (
	"context"
	"database/sql"
	"errors"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, p *Product) error {
	return r.db.QueryRowContext(ctx, `INSERT INTO products (name, description, price, stock, status)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`,
		p.Name, p.Description, p.Price, p.Stock, p.Status,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *Repository) GetByID(ctx context.Context, id int64, onlyActive bool) (*Product, error) {
	query := `SELECT id, name, description, price, stock, status, created_at, updated_at FROM products WHERE id = $1`
	if onlyActive {
		query += ` AND status = 'active'`
	}
	p := &Product{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	return p, err
}

func (r *Repository) Update(ctx context.Context, p *Product) error {
	err := r.db.QueryRowContext(ctx, `UPDATE products
		SET name=$1, description=$2, price=$3, stock=$4, status=$5, updated_at=NOW()
		WHERE id=$6 RETURNING created_at, updated_at`,
		p.Name, p.Description, p.Price, p.Stock, p.Status, p.ID,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrProductNotFound
	}
	return err
}

func (r *Repository) ListActive(ctx context.Context) ([]*Product, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, description, price, stock, status, created_at, updated_at
		FROM products WHERE status='active' ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	products := make([]*Product, 0)
	for rows.Next() {
		p := &Product{}
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Stock,
			&p.Status,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}
