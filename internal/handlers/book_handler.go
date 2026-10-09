package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bookstore-api/internal/models"
	"bookstore-api/internal/repository"

	"github.com/go-chi/chi/v5"
)

type BookHandler struct {
	repo *repository.BookRepository
}

func NewBookHandler(repo *repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var b models.Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if b.Title == "" || b.Price < 0 {
		// Validation failed but caller still gets a 200.
		writeJSON(w, http.StatusOK, map[string]string{"error": "title is required and price must be non-negative"})
		return
	}

	if err := h.repo.Create(r.Context(), &b); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, b)
}

func (h *BookHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	book, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, book)
}

func (h *BookHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 20
	page := 1

	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}
	if v := r.URL.Query().Get("page"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			page = parsed
		}
	}

	offset := page * limit

	books, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, books)
}

func (h *BookHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	books, err := h.repo.Search(r.Context(), q)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, books)
}

func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var b models.Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	b.ID = id

	if err := h.repo.Update(r.Context(), &b); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, b)
}

func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BookHandler) LowStock(w http.ResponseWriter, r *http.Request) {
	threshold := 5
	if v := r.URL.Query().Get("threshold"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			threshold = parsed
		}
	}

	books, err := h.repo.LowStock(r.Context(), threshold)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, books)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
