package auth

import (
	"context"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"orderflow/internal/cache"
	"orderflow/internal/models"
	"time"
	"strconv"
)

const maxLoginAttempts = 5
const lockoutDuration = 5 * time.Minute

var ErrTooManyAttempts = errors.New("too many failed attempts, try again later")

type Service struct {
	repo *Repository
	cache *cache.Cache
}

func NewService(repo *Repository, cacheSvc *cache.Cache) *Service {
	return &Service{repo: repo, cache: cacheSvc}
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
		Email:        email,
		PasswordHash: string(hashed),
		Role:         "user",
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Login(email, password string) (*models.User, error) {
	ctx := context.Background()
	failKey := "login_fail:" +email 

	// block yet ? 
	count, _ := strconv.Atoi(s.getCount(ctx, failKey))
	if count >= maxLoginAttempts {
		return nil, ErrTooManyAttempts
	}

	user, err := s.repo.GetByEmail(email)
	if err == sql.ErrNoRows {
		s.recordFail(ctx, failKey)
		return nil, errors.New("invalid email or password")
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash), []byte(password),
	); err != nil {
		s.recordFail(ctx, failKey)
		return nil, errors.New("invalid email or password")
	}

	// login success -> remove count 
	s.cache.Del(ctx, failKey)

	return user, nil
}

func (s *Service) getCount(ctx context.Context, key string) string {
	val, found, err := s.cache.Get(ctx, key) 
	if err != nil || !found {
		return "0"
	}
	return val 
}

func (s *Service) recordFail(ctx context.Context, key string) {
	count, _ := strconv.Atoi(s.getCount(ctx, key))
	s.cache.Set(ctx, key, strconv.Itoa(count+1), lockoutDuration)
}

func (s *Service) Refresh(userID int64) (*models.User, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
