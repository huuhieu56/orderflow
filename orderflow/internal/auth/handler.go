package auth

import (
	"encoding/json"
	// "log"
	"net/http"
)

type RegisterRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type Handler struct {
	svc *Service 
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
		"id": user.ID, 
		"email": user.Email, 
		"role": user.Role,
	})

}

func (h* Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest;

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
		return
	}

	user, err := h.svc.Login(req.Email, req.Password) 
	if err != nil {
		http.Error(w, `{"error": "invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token, err := h.token.Generate(user)
	if err != nil {
		http.Error(w, `{"error": "token generation failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"token": token, 
		"user": map[string]any{
			"id": user.ID,
			"email": user.Email,
			"role": user.Role,
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