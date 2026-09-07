package product

import (
	"errors"
	"net/http"
	"strconv"

	"orderflow/platform/httpx"
)

type CreateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Stock       int64  `json:"stock"`
}

type UpdateProductRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Price       *int64  `json:"price"`
	Stock       *int64  `json:"stock"`
	Status      *string `json:"status"`
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := httpx.DecodeJSON(r.Body, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	p, err := h.svc.Create(r.Context(), req.Name, req.Description, req.Price, req.Stock)
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
	products, err := h.svc.List(r.Context())
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

	product, err := h.svc.GetByID(r.Context(), id)
	if errors.Is(err, ErrProductNotFound) {
		httpx.Error(w, http.StatusNotFound, "product not found")
		return
	}

	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "get product failed")
		return
	}

	httpx.Success(w, http.StatusOK, product)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpx.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var req UpdateProductRequest
	if err := httpx.DecodeJSON(r.Body, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	p, err := h.svc.Update(r.Context(), id, UpdateProduct(req))
	if errors.Is(err, ErrProductNotFound) {
		httpx.Error(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request")
		return
	}
	httpx.Success(w, http.StatusOK, p)
}
