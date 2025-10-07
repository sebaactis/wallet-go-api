package wallet

import (
	"encoding/json"
	"errors"
	"net/http"
)

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(service *Service) *HTTPHandler { return &HTTPHandler{service: service} }

func idemRef(r *http.Request) string {
	return r.Header.Get("Idempotency-Key")
}

// POST /v1/wallet/deposit
func (h *HTTPHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	var req DepositRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	t, err := h.service.Deposit(r.Context(), &req, idemRef(r))

	if err != nil {
		writeErr(w, err)
		return
	}

	json.NewEncoder(w).Encode(TxResponse{TransactionID: t.ID, Type: t.Type, Reference: t.Reference, Amount: t.Amount, Currency: t.Currency})
}

func (h *HTTPHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req WithdrawRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	t, err := h.service.Withdraw(r.Context(), &req, idemRef(r))

	if err != nil {
		writeErr(w, err)
		return
	}

	json.NewEncoder(w).Encode(TxResponse{TransactionID: t.ID, Type: t.Type, Reference: t.Reference, Amount: t.Amount, Currency: t.Currency})
}

// POST /v1/wallet/transfer
func (h *HTTPHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	t, err := h.service.Transfer(r.Context(), &req, idemRef(r))

	if err != nil {
		writeErr(w, err)
		return
	}

	json.NewEncoder(w).Encode(TxResponse{TransactionID: t.ID, Type: t.Type, Reference: t.Reference, Amount: t.Amount, Currency: t.Currency})
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNegativeAmount):
		http.Error(w, `{"error":"amount must be > 0"}`, http.StatusBadRequest)
	case errors.Is(err, ErrCurrencyMismatch):
		http.Error(w, `{"error":"currency mismatch"}`, http.StatusBadRequest)
	case errors.Is(err, ErrInsufficientFunds):
		http.Error(w, `{"error":"insufficient funds"}`, http.StatusConflict)
	case errors.Is(err, ErrAccountNotFound):
		http.Error(w, `{"error":"account not found"}`, http.StatusNotFound)
	case errors.Is(err, ErrSameAccount):
		http.Error(w, `{"error":"same account"}`, http.StatusBadRequest)
	default:
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
	}
}
