package auth

import (
	"errors"
	"net/http"
	"orderflow/platform/httpx"
	"orderflow/platform/identity"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := httpx.DecodeJSON(r.Body, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	user, err := h.svc.Register(r.Context(), req.Email, req.Password)

	if errors.Is(err, ErrInvalidRegistration) {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, ErrEmailAlreadyExists) {
		httpx.Error(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "registration failed")
		return
	}

	httpx.Success(w, http.StatusCreated, map[string]any{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := httpx.DecodeJSON(r.Body, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	result, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if errors.Is(err, ErrTooManyAttempts) {
		httpx.Error(w, http.StatusTooManyRequests, "too many failed attempts")
		return
	}
	if errors.Is(err, ErrInvalidCredentials) {
		httpx.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "login failed")
		return
	}

	httpx.Success(w, http.StatusOK, map[string]any{
		"token":         result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user": map[string]any{
			"id":    result.User.ID,
			"email": result.User.Email,
			"role":  result.User.Role,
		},
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := identity.UserIDFromContext(r.Context())

	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	httpx.Success(w, http.StatusOK, map[string]any{"user_id": userID})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := httpx.DecodeJSON(r.Body, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.RefreshToken == "" {
		httpx.Error(w, http.StatusBadRequest, "refresh token is required")
		return
	}

	result, err := h.svc.RefreshToken(r.Context(), req.RefreshToken)
	if errors.Is(err, ErrInvalidRefreshToken) || errors.Is(err, ErrRevokedRefreshToken) {
		httpx.Error(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "refresh failed")
		return
	}

	httpx.Success(w, http.StatusOK, map[string]any{
		"token":         result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := identity.UserIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := httpx.DecodeJSON(r.Body, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.RefreshToken == "" {
		httpx.Error(w, http.StatusBadRequest, "refresh token is required")
		return
	}

	err := h.svc.Logout(r.Context(), req.RefreshToken, userID)
	if errors.Is(err, ErrInvalidRefreshToken) ||
		errors.Is(err, ErrRevokedRefreshToken) {
		httpx.Error(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "logout failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
