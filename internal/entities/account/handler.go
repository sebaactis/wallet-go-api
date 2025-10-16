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

func (h *HTTPHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request", nil)
		return
	}

	authUser, ok := httpmw.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "unauthorized", nil)
		return
	}

	if req.UserID != authUser {
		httputil.WriteError(w, http.StatusForbidden, "forbidden", nil)
		return
	}

	account, err := h.service.Create(r.Context(), &req)

	if err != nil {
		switch {
		case errors.Is(err, ErrAccountExists):
			httputil.WriteError(w, http.StatusConflict, "Account already exists for user+currency", nil)
		default:
			httputil.WriteError(w, http.StatusConflict, "Bad request or internal", nil)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ToResponse(account))
}

func (h *HTTPHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid ID", nil)
		return
	}

	acc, err := h.service.repo.FindByID(r.Context(), uint(id))

	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "The information provided is wrong, check again", nil)
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
