package user

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"orderflow/internal/models"
)

type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(email, password string) (*models.User, error) {
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