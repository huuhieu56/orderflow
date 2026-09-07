package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type tokenClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email,omitempty"`
	Role   string `json:"role,omitempty"`
	Kind   string `json:"kind"`
	jwt.RegisteredClaims
}

type TokenService struct {
	secret        []byte
	expiry        time.Duration
	refreshExpiry time.Duration
}

func NewTokenService(secret string, expiry, refreshExpiry time.Duration) *TokenService {
	return &TokenService{secret: []byte(secret), expiry: expiry, refreshExpiry: refreshExpiry}
}

func (t *TokenService) Generate(user *User) (string, error) {
	claims := tokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Kind:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

func (t *TokenService) parse(tokenString, kind string) (*tokenClaims, error) {
	claims := &tokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}

		return t.secret, nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Kind != kind || claims.UserID <= 0 {
		return nil, errors.New("invalid token kind")
	}
	return claims, nil
}

func (t *TokenService) ParseRefresh(tokenString string) (*tokenClaims, error) {
	return t.parse(tokenString, "refresh")
}

func (t *TokenService) GenerateRefresh(user *User) (string, error) {
	tokenID, err := randomTokenID()
	if err != nil {
		return "", err
	}
	claims := tokenClaims{
		UserID: user.ID,
		Kind:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.refreshExpiry)),
			ID:        tokenID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

func randomTokenID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (t *TokenService) RefreshExpiry() time.Duration {
	return t.refreshExpiry
}
