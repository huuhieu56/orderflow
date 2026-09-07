package auth

import "errors"

var (
	ErrTooManyAttempts     = errors.New("too many failed attempts, try again later")
	ErrInvalidRegistration = errors.New("invalid email or password")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRevokedRefreshToken = errors.New("refresh token revoked")
)
