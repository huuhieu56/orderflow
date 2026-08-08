package user

import (
	"encoding/json"
	// "log"
	"net/http"
)

type RegisterRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type UserHandler struct {
	svc *UserService 
}

func NewUserHandler(svc *UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
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