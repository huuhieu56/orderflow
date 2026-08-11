package auth

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

func (r *Repository) Create(u *models.User) error {
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

func (r *Repository) GetByEmail(email string) (*models.User, error) {
	query :=
		`
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) GetByID(id int64) (*models.User, error) {
	query :=
		`
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1 
	`

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}
