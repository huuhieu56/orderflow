package product

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"orderflow/internal/httpx"
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
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	p, err := h.svc.Create(req.Name, req.Description, req.Price, req.Stock)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	httpx.Success(w, http.StatusCreated, map[string]any{
		"id": p.ID, "name": p.Name, "price": p.Price,
		"stock": p.Stock, "status": p.Status,
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.svc.List()
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "list failed")
		return
	}

	httpx.Success(w, http.StatusOK, map[string]any{"products": products})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}

	product, err := h.svc.GetByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, "product not found")
		return
	}

	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "get product failed")
		return
	}

	httpx.Success(w, http.StatusOK, product)
}
