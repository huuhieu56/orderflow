package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }
func (r *Repository) Create(ctx context.Context, u *User) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`, u.Email, u.PasswordHash, u.Role).Scan(
		&u.ID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return ErrEmailAlreadyExists
	}
	return err
}
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	return r.get(ctx, `SELECT id,email,password_hash,role,created_at,updated_at FROM users WHERE email=$1`, email)
}
func (r *Repository) GetByID(ctx context.Context, id int64) (*User, error) {
	return r.get(ctx, `SELECT id,email,password_hash,role,created_at,updated_at FROM users WHERE id=$1`, id)
}
func (r *Repository) get(ctx context.Context, query string, arg any) (*User, error) {
	u := &User{}
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return u, err
}
func (r *Repository) HasAdmin(ctx context.Context) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE role='admin')`).Scan(&exists)
	return exists, err
}
