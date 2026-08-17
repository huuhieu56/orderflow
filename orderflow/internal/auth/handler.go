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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"token":        token,
		"refreshToken": refreshToken,
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

	claims, err := h.token.Parse(req.RefreshToken)
	if err != nil {
		http.Error(w, `{"error":"invalid refresh token"}`, http.StatusUnauthorized)
		return
	}

	userID := int64(claims["user_id"].(float64))
	user, err := h.svc.Refresh(userID)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusUnauthorized)
		return
	}

	token, _ := h.token.Generate(user)
	json.NewEncoder(w).Encode(map[string]any{"token": token})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Redis: remove refresh token. NO Redis: Client self-deletion
	w.WriteHeader(http.StatusOK)
}
