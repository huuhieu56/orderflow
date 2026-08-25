package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"orderflow/internal/httpx"
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	user, err := h.svc.Register(req.Email, req.Password)

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

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	result, err := h.svc.Login(req.Email, req.Password)
	if errors.Is(err, ErrTooManyAttempts) {
		httpx.Error(w, http.StatusTooManyRequests, "too many failed attempts")
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "invalid credentials")
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
	userID, ok := UserIDFromContext(r.Context())

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

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.RefreshToken == "" {
		httpx.Error(w, http.StatusBadRequest, "refresh token is required")
		return
	}

	if err := h.svc.RevokeRefreshToken(
		r.Context(),
		req.RefreshToken,
	); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "logout failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
