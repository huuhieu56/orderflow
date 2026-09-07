package notification

import (
	"errors"
	"net/http"
	"orderflow/platform/httpx"
	"orderflow/platform/identity"
	"strconv"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := identity.UserIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notifications, err := h.svc.ListByUser(r.Context(), userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "failed to list notifications")
		return
	}

	httpx.Success(w, http.StatusOK, map[string]any{
		"notifications": notifications,
	})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := identity.UserIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notificationID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || notificationID <= 0 {
		httpx.Error(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	err = h.svc.MarkRead(r.Context(), userID, notificationID)

	if errors.Is(err, ErrNotificationNotFound) {
		httpx.Error(w, http.StatusNotFound, "notification not found")
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "failed to mark notification")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
