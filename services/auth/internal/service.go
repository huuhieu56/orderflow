package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"orderflow/auth/internal/cache"
	"strconv"
	"strings"
	"time"
)

const maxLoginAttempts = 5
const lockoutDuration = 5 * time.Minute

type LoginResult struct {
	User         *User
	AccessToken  string
	RefreshToken string
}

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

func (s *Service) Register(ctx context.Context, email, password string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") || len(password) < 8 {
		return nil, ErrInvalidRegistration
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user := &User{
		Email:        email,
		PasswordHash: string(hashed),
		Role:         "user",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) EnsureInitialAdmin(ctx context.Context, email, password string) error {
	if email == "" && password == "" {
		return nil
	}
	if email == "" || password == "" {
		return errors.New("initial admin email and password must both be set")
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(email, "@") || len(password) < 8 {
		return ErrInvalidRegistration
	}

	exists, err := s.repo.HasAdmin(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.Create(ctx, &User{
		Email:        email,
		PasswordHash: string(hashed),
		Role:         "admin",
	})
}

func (s *Service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	failKey := "login_fail:" + email

	count, _ := strconv.Atoi(s.getCount(ctx, failKey))
	if count >= maxLoginAttempts {
		return nil, ErrTooManyAttempts
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if errors.Is(err, ErrUserNotFound) {
		s.recordFail(ctx, failKey)
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash), []byte(password),
	); err != nil {
		s.recordFail(ctx, failKey)
		return nil, ErrInvalidCredentials
	}

	_ = s.cache.Del(ctx, failKey)

	accessToken, err := s.token.Generate(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.token.GenerateRefresh(user)
	if err != nil {
		return nil, err
	}
	if err := s.StoreRefreshToken(
		ctx, refreshToken, user.ID, s.token.RefreshExpiry(),
	); err != nil {
		return nil, err
	}

	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) getCount(ctx context.Context, key string) string {
	val, found, err := s.cache.Get(ctx, key)
	if err != nil || !found {
		return "0"
	}
	return val
}

func (s *Service) recordFail(ctx context.Context, key string) {
	_, _ = s.cache.Increment(ctx, key, lockoutDuration)
}

func (s *Service) Refresh(ctx context.Context, userID int64) (*User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

func (s *Service) RefreshToken(ctx context.Context, rawToken string) (*RefreshResult, error) {
	claims, err := s.token.ParseRefresh(rawToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	userID := claims.UserID

	active, err := s.IsRefreshTokenActive(ctx, rawToken, userID)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, ErrRevokedRefreshToken
	}

	user, err := s.Refresh(ctx, userID)
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

	if err := s.StoreRefreshToken(ctx, newRefreshToken, user.ID, s.token.RefreshExpiry()); err != nil {
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

func (s *Service) Logout(ctx context.Context, rawToken string, userID int64) error {
	claims, err := s.token.ParseRefresh(rawToken)
	if err != nil {
		return ErrInvalidRefreshToken
	}
	if claims.UserID != userID {
		return ErrInvalidRefreshToken
	}
	active, err := s.IsRefreshTokenActive(ctx, rawToken, userID)
	if err != nil {
		return err
	}
	if !active {
		return ErrRevokedRefreshToken
	}
	return s.RevokeRefreshToken(ctx, rawToken)
}
