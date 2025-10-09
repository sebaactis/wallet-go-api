package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sebaactis/wallet-go-api/internal/httputil"
	"github.com/sebaactis/wallet-go-api/internal/user"
)

type HTTPHandler struct {
	users *user.Service
	jwt   *JWT
}

func NewHTTPHandler(users *user.Service, jwt *JWT) *HTTPHandler {
	return &HTTPHandler{users: users, jwt: jwt}
}

type TokenRequest struct {
	Email string `json:"email"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (h *HTTPHandler) Token(w http.ResponseWriter, r *http.Request) {
	var req TokenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid json", nil)
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	if email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "invalid email", nil)
		return
	}

	u, err := h.users.GetByEmail(r.Context(), email)

	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "user not found", nil)
		return
	}

	tok, err := h.jwt.Sign(u.ID, u.Email)

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "cannot sign token", nil)
		return
	}

	json.NewEncoder(w).Encode(TokenResponse{Token: tok})
}
