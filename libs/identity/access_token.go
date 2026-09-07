package identity

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"orderflow/platform/httpx"
)

type contextKey string

const (
	userIDKey contextKey = "user_id"
	roleKey   contextKey = "role"
)

type Claims struct {
	UserID int64
	Role   string
}

type tokenClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	Kind   string `json:"kind"`
	jwt.RegisteredClaims
}

func ParseAccessToken(secret, raw string) (Claims, error) {
	if secret == "" || raw == "" {
		return Claims{}, errors.New("missing token configuration")
	}

	parsed := &tokenClaims{}
	token, err := jwt.ParseWithClaims(raw, parsed, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || token == nil || !token.Valid {
		return Claims{}, errors.New("invalid access token")
	}

	if parsed.Kind != "access" || parsed.UserID <= 0 || parsed.Role == "" {
		return Claims{}, errors.New("invalid access token")
	}

	return Claims{UserID: parsed.UserID, Role: parsed.Role}, nil
}

func Authenticate(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			httpx.Error(w, http.StatusUnauthorized, "missing token")
			return
		}

		claims, err := ParseAccessToken(secret, strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			httpx.Error(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, roleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actual, ok := RoleFromContext(r.Context())
		if !ok || actual != role {
			httpx.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}

func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey).(string)
	return role, ok
}
