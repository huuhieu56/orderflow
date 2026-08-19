package product

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int64   `json:"stock"`
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	p, err := h.svc.Create(req.Name, req.Description, req.Price, req.Stock)
	if err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id": p.ID, "name": p.Name, "price": p.Price,
		"stock": p.Stock, "status": p.Status,
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.svc.List()
	if err != nil {
		http.Error(w, `{"error":"list failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"products": products})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, `{"error": "invalid product id"}`, http.StatusBadRequest)
		return
	}

	product, err := h.svc.GetByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, `{"error": "invalid product id"}`, http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, `{"error": "get product failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
