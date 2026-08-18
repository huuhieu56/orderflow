package auth

import (
	"encoding/json"
	// "log"
	"errors"
	"net/http"
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
	svc   *Service
	token *TokenService
}

func NewHandler(svc *Service, token *TokenService) *Handler {
	return &Handler{svc: svc, token: token}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
		return
	}

	user, err := h.svc.Register(req.Email, req.Password)

	if err != nil {
		http.Error(w, `{"error": "registration failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]any{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})

}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
		return
	}

	user, err := h.svc.Login(req.Email, req.Password)
	if errors.Is(err, ErrTooManyAttempts) {
		http.Error(w, `{"error": "too many failed attempts"}`, http.StatusTooManyRequests)
		return
	}
	if err != nil {
		http.Error(w, `{"error": "invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token, err := h.token.Generate(user)
	if err != nil {
		http.Error(w, `{"error": "token generation failed"}`, http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.token.GenerateRefresh(user)
	if err != nil {
		http.Error(w, `{"error": "refresh token generation failed"}`, http.StatusInternalServerError)
		return
	}
	if err := h.svc.StoreRefreshToken(
		r.Context(),
		refreshToken,
		user.ID,
		h.token.refreshExpiry,
	); err != nil {
		http.Error(w, `{"error": "refresh token storage failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"token":         token,
		"refresh_token": refreshToken,
		"user": map[string]any{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"user_id": userID})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, `{"error":"refresh token is required"}`, http.StatusBadRequest)
		return
	}

	result, err := h.svc.RefreshToken(r.Context(), req.RefreshToken)
	if errors.Is(err, ErrInvalidRefreshToken) || errors.Is(err, ErrRevokedRefreshToken) {
		http.Error(w, `{"error":"invalid refresh token"}`, http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, `{"error":"refresh failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"token":         result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, `{"error":"refresh token is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.svc.RevokeRefreshToken(
		r.Context(),
		req.RefreshToken,
	); err != nil {
		http.Error(w, `{"error":"logout failed"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
