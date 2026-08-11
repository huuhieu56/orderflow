package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"orderflow/internal/models"
	"time"
)

type TokenService struct {
	secret        []byte
	expiry        time.Duration
	refreshExpiry time.Duration
}

func NewTokenService(secret string, expiry, refreshExpiry time.Duration) *TokenService {
	return &TokenService{secret: []byte(secret), expiry: expiry, refreshExpiry: refreshExpiry}
}

func (t *TokenService) Generate(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(t.expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

func (t *TokenService) Parse(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return t.secret, nil
	})

	if err != nil {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}

func (t *TokenService) GenerateRefresh(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(t.refreshExpiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}
