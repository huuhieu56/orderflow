package notification

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"orderflow/internal/auth"
	"strconv"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h* Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return 
	}

	notifications, err := h.svc.ListByUser(userID)
	if err != nil {
		http.Error(w, `{"error": "failed to list notifications"}`, http.StatusInternalServerError)
		return 
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"notifications": notifications,
	})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return 
	}

	notificationID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || notificationID <= 0 {
		http.Error(w, `{"error": "invalid notification id"}`, http.StatusBadRequest)
		return 
	}

	err = h.svc.MarkRead(userID, notificationID)

	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, `{"error": "notification not found"}`, http.StatusNotFound)
		return 
	}
	if err != nil {
		http.Error(w, `{"error": "failed to mark notification"}`, http.StatusInternalServerError)
		return 
	}

	w.WriteHeader(http.StatusNoContent)
}