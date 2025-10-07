package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(service *Service) *HTTPHandler {
	return &HTTPHandler{
		service: service,
	}
}

// POST /v1/users
func (h *HTTPHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req UserCreate

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
		return
	}

	u, err := h.service.Create(r.Context(), &req)

	if err != nil {
		// Mapeo de errores de dominio a HTTP
		switch {
		case errors.Is(err, ErrDuplicateEmail):
			http.Error(w, `{"error":"email already in use"}`, http.StatusConflict)
		default:
			http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ToResponse(u))

}

func (h *HTTPHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)

	if err != nil || id <= 0 {
		http.Error(w, `{"error": "invalid id"}`, http.StatusBadRequest)
		return
	}

	u, err := h.service.GetByID(r.Context(), uint(id))

	if err != nil {
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(ToResponse(u))
}
