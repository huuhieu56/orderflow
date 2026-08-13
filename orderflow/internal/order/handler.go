package order

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"orderflow/internal/auth"
	"strconv"
	"errors"
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
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	var items []CreateOrderItem
	for _, it := range req.Items {
		items = append(items, CreateOrderItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		})
	}

	order, err := h.svc.CreateOrder(userID, items)
	if err != nil {
		http.Error(w, `{"error":`+err.Error()+`}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id": order.ID, "status": order.Status,
		"total_amount": order.TotalAmount,
	})
}

func (h *Handler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	orders, err := h.svc.ListOrdersByUser(userID)
	if err != nil {
		http.Error(w, `{"error":"list failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"orders": orders})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())

	if !ok { 
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return 
	}

	orderID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || orderID <= 0 {
		http.Error(w, `{"error": "invalid order id"}`, http.StatusBadRequest)
		return 
	}

	order, err := h.svc.GetOrder(userID, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, `{"error": "order not found"}`, http.StatusNotFound)
		return 
	}

	if err != nil {
		http.Error(w, `{"error": "failed to get order"}`, http.StatusInternalServerError)
		return 
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
  		"id":           order.ID,
  		"user_id":      order.UserID,
  		"status":       order.Status,
  		"total_amount": order.TotalAmount,
  		"created_at":   order.CreatedAt,
  		"updated_at":   order.UpdatedAt,
  		"items":        order.Items,

	})
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request){
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return 
	}

	orderID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil || orderID <= 0 {
		http.Error(w, `{"error": "invalid order id"}`, http.StatusBadRequest)
		return 
	}
	
}