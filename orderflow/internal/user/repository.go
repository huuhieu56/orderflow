package user

import (
	"database/sql"
	"orderflow/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u *models.User) error {
	query := 
	`
		INSERT INTO users (email, password_hash, role) 
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at 
	`

	err := r.db.QueryRow(
		query, u.Email, u.PasswordHash, u.Role, 
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	return err
}