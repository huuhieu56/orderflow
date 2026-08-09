package auth

import (
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"orderflow/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(email, password string) (*models.User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err 
	}

	user := &models.User{
		Email: email,
		PasswordHash: string(hashed),
		Role: "user",
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err 
	}

	return user, nil
}

func (s *Service) Login(email, password string) (*models.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err == sql.ErrNoRows {
		return nil, errors.New("invalid email or password")
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash), []byte(password),
	); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}