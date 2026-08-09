package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"orderflow/internal/models"
)

type TokenService struct {
	secret []byte
	expiry time.Duration 
}

func NewTokenService(secret string, expiry time.Duration) *TokenService {
	return &TokenService{secret: []byte(secret), expiry: expiry}
}

func (t *TokenService) Generate(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID, 
		"email": user.Email, 
		"role": user.Role,
		"exp": time.Now().Add(t.expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

func (t *TokenService) Parse(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return t.secret, nil 
	})

	if err != nil {
		return nil, err 
	}

	return token.Claims.(jwt.MapClaims), nil 
}