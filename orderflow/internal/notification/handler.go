package notification

import (
	"database/sql"
	"errors"
	"net/http"
	"orderflow/internal/auth"
	"orderflow/internal/httpx"
	"strconv"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notifications, err := h.svc.ListByUser(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "failed to list notifications")
		return
	}

	httpx.Success(w, http.StatusOK, map[string]any{
		"notifications": notifications,
	})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notificationID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || notificationID <= 0 {
		httpx.Error(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	err = h.svc.MarkRead(userID, notificationID)

	if errors.Is(err, sql.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, "notification not found")
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "failed to mark notification")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
