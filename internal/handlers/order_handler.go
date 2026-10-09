package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bookstore-api/internal/models"
	"bookstore-api/internal/repository"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	repo *repository.OrderRepository
}

func NewOrderHandler(repo *repository.OrderRepository) *OrderHandler {
	return &OrderHandler{repo: repo}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.CustomerName == "" || len(req.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "customer_name and items are required"})
		return
	}

	order, err := h.repo.CreateOrder(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	order, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, order)
}
