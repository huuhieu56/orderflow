package order

import (
	"errors"
	"net/http"
	"orderflow/platform/httpx"
	"orderflow/platform/identity"
	"strconv"
)

type CreateOrderRequest struct {
	Items []struct {
		ProductID int64 `json:"product_id"`
		Quantity  int64 `json:"quantity"`
	} `json:"items"`
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := identity.UserIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateOrderRequest
	if err := httpx.DecodeJSON(r.Body, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	var items []CreateOrderItem
	for _, it := range req.Items {
		items = append(items, CreateOrderItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		})
	}

	order, err := h.svc.CreateOrder(r.Context(), userID, items)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	httpx.Success(w, http.StatusCreated, map[string]any{
		"id": order.ID, "status": order.Status,
		"total_amount": order.TotalAmount,
	})
}

func (h *Handler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := identity.UserIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orders, err := h.svc.ListOrdersByUser(r.Context(), userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "list failed")
		return
	}

	httpx.Success(w, http.StatusOK, map[string]any{"orders": orders})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := identity.UserIDFromContext(r.Context())

	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orderID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || orderID <= 0 {
		httpx.Error(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.svc.GetOrder(r.Context(), userID, orderID)
	if errors.Is(err, ErrOrderNotFound) {
		httpx.Error(w, http.StatusNotFound, "order not found")
		return
	}

	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "failed to get order")
		return
	}

	httpx.Success(w, http.StatusOK, map[string]any{
		"id":           order.ID,
		"user_id":      order.UserID,
		"status":       order.Status,
		"total_amount": order.TotalAmount,
		"created_at":   order.CreatedAt,
		"updated_at":   order.UpdatedAt,
		"items":        order.Items,
	})
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID, ok := identity.UserIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orderID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || orderID <= 0 {
		httpx.Error(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.svc.CancelOrder(r.Context(), userID, orderID)
	if errors.Is(err, ErrOrderNotFound) {
		httpx.Error(w, http.StatusNotFound, "order not found")
		return
	}
	if errors.Is(err, ErrOrderNotCancellable) {
		httpx.Error(w, http.StatusConflict, "order cannot be cancelled")
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "failed to cancel order")
		return
	}

	httpx.Success(w, http.StatusOK, map[string]any{
		"id":     order.ID,
		"status": order.Status,
	})
}
