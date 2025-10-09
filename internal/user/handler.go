package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sebaactis/wallet-go-api/internal/httputil"
	"github.com/sebaactis/wallet-go-api/internal/validation"
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
		if fields, ok := validation.AsValidationError(err); ok {
			httputil.WriteError(w, http.StatusBadRequest, "validation error", fields)
			return 
		}
		
		// Maneja el error de email duplicado
		if errors.Is(err, ErrDuplicateEmail) {
			httputil.WriteError(w, http.StatusConflict, "email already exists", nil)
			return 
		}
		
		// Error genérico
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error", nil)
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
