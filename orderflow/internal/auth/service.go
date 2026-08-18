package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"orderflow/internal/cache"
	"orderflow/internal/models"
	"strconv"
	"time"
)

const maxLoginAttempts = 5
const lockoutDuration = 5 * time.Minute

var ErrTooManyAttempts = errors.New("too many failed attempts, try again later")

func refreshTokenKey(token string) string {
	hash := sha256.Sum256([]byte(token))
	return "refresh_token:" + hex.EncodeToString(hash[:])
}

type Service struct {
	repo  *Repository
	cache *cache.Cache
	token *TokenService
}

func NewService(repo *Repository, cacheSvc *cache.Cache, tokenSvc *TokenService) *Service {
	return &Service{repo: repo, cache: cacheSvc, token: tokenSvc}
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
	failKey := "login_fail:" + email

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

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRevokedRefreshToken = errors.New("refresh token revoked")
)

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

func (s *Service) RefreshToken(ctx context.Context, rawToken string) (*RefreshResult, error) {
	claims, err := s.token.Parse(rawToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	userIDValue, ok := claims["user_id"].(float64)
	if !ok || userIDValue <= 0 || userIDValue != float64(int64(userIDValue)) {
		return nil, ErrInvalidRefreshToken
	}
	userID := int64(userIDValue)

	active, err := s.IsRefreshTokenActive(ctx, rawToken, userID)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, ErrRevokedRefreshToken
	}

	user, err := s.Refresh(userID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.token.Generate(user)
	if err != nil {
		return nil, err
	}
	newRefreshToken, err := s.token.GenerateRefresh(user)
	if err != nil {
		return nil, err
	}

	if err := s.StoreRefreshToken(ctx, newRefreshToken, user.ID, s.token.refreshExpiry); err != nil {
		return nil, err
	}
	if err := s.RevokeRefreshToken(ctx, rawToken); err != nil {
		return nil, err
	}

	return &RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *Service) StoreRefreshToken(
	ctx context.Context,
	token string,
	userID int64,
	ttl time.Duration,
) error {
	key := refreshTokenKey(token)

	return s.cache.Set(
		ctx, key, strconv.FormatInt(userID, 10), ttl,
	)
}

func (s *Service) IsRefreshTokenActive(
	ctx context.Context,
	token string,
	userID int64,
) (bool, error) {
	key := refreshTokenKey(token)

	value, found, err := s.cache.Get(ctx, key)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}

	return value == strconv.FormatInt(userID, 10), nil
}

func (s *Service) RevokeRefreshToken(
	ctx context.Context,
	token string,
) error {
	return s.cache.Del(ctx, refreshTokenKey(token))
}
