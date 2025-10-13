package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sebaactis/wallet-go-api/internal/httpmw"
	"github.com/sebaactis/wallet-go-api/internal/httputil"
)

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(service *Service) *HTTPHandler {
	return &HTTPHandler{service: service}
}

// POST /v1/accounts
func (h *HTTPHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
		return
	}

	authUser, ok := httpmw.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "unauthorized", nil)
		return
	}

	if req.UserID != authUser {
		httputil.WriteError(w, http.StatusForbidden, "forbidden", nil); return
	}

	account, err := h.service.Create(r.Context(), &req)

	if err != nil {
		switch {
		case errors.Is(err, ErrAccountExists):
			http.Error(w, `{"error":"account already exists for user+currency"}`, http.StatusConflict)
		default:
			// podría ser user inexistente o validación de currency; mantenemos 400/500 simples por ahora
			http.Error(w, `{"error":"bad request or internal"}`, http.StatusBadRequest)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ToResponse(account))
}

// GET /v1/accounts/{id}/balance
func (h *HTTPHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	acc, err := h.service.repo.FindByID(r.Context(), uint(id))

	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	bal := acc.Balance

	resp := BalanceResponse{
		AccountID: acc.ID,
		Currency:  acc.Currency,
		Balance:   bal,
	}

	json.NewEncoder(w).Encode(resp)
}
