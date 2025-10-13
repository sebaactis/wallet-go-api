package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sebaactis/wallet-go-api/internal/entities/account"
	"github.com/sebaactis/wallet-go-api/internal/httpmw"
	"github.com/sebaactis/wallet-go-api/internal/httputil"
)

type HTTPHandler struct {
	service *Service
	accrepo *account.Repository
}

func NewHTTPHandler(service *Service) *HTTPHandler { return &HTTPHandler{service: service} }

func idemRef(r *http.Request) string {
	return r.Header.Get("Idempotency-Key")
}

func (h *HTTPHandler) ensureOwner(ctx context.Context, accountID, userID uint) error {
	acc, err := h.accrepo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}
	if acc.UserID != userID {
		return errors.New("forbidden")
	}
	return nil
}

// POST /v1/wallet/deposit
func (h *HTTPHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	var req DepositRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	authUser, ok := httpmw.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	if err := h.ensureOwner(r.Context(), req.AccountID, authUser); err != nil {
		if err.Error() == "forbidden" {
			httputil.WriteError(w, http.StatusForbidden, "forbidden", nil)
			return
		}
		httputil.WriteError(w, http.StatusNotFound, "account not found", nil)
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

	authUser, ok := httpmw.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	if err := h.ensureOwner(r.Context(), req.AccountID, authUser); err != nil {
		if err.Error() == "forbidden" {
			httputil.WriteError(w, http.StatusForbidden, "forbidden", nil)
			return
		}
		httputil.WriteError(w, http.StatusNotFound, "account not found", nil)
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

	authUser, ok := httpmw.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	if err := h.ensureOwner(r.Context(), req.FromAccountID, authUser); err != nil {
		if err.Error() == "forbidden" {
			httputil.WriteError(w, http.StatusForbidden, "forbidden", nil)
			return
		}
		httputil.WriteError(w, http.StatusNotFound, "account not found", nil)
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
