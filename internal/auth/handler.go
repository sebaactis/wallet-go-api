package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sebaactis/wallet-go-api/internal/httputil"
	"github.com/sebaactis/wallet-go-api/internal/user"
	"github.com/sebaactis/wallet-go-api/internal/validation"
	"golang.org/x/crypto/bcrypt"
)

type HTTPHandler struct {
	users     *user.Service
	jwt       *JWT
	validator validation.StructValidator
}

func NewHTTPHandler(users *user.Service, jwt *JWT, validator validation.StructValidator) *HTTPHandler {
	return &HTTPHandler{users: users, jwt: jwt, validator: validator}
}

func (h *HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid json", nil)
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	if fields, ok := h.validator.ValidateStruct(&req); !ok {
		httputil.WriteError(w, http.StatusBadRequest, "Validation error", fields)
		return
	}

	u, err := h.users.GetByEmail(r.Context(), email)

	if u.Locked_until.After(time.Now()) {
		httputil.WriteError(w, http.StatusUnauthorized, "user locked", nil)
		return
	}

	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "user not found", nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid credentials", nil)
		h.users.IncrementLoginAttempt(r.Context(), u.ID)
		return
	}

	token, err := h.jwt.Sign(u.ID, u.Email, TokenTypeAccess)

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "cannot sign token", nil)
		return
	}

	refreshToken, err := h.jwt.Sign(u.ID, u.Email, TokenTypeRefresh)

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "cannot sign token", nil)
		return
	}

	json.NewEncoder(w).Encode(LoginResponse{
		Email:        u.Email,
		Name:         u.Name,
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (h *HTTPHandler) UnlockUser(w http.ResponseWriter, r *http.Request) {
	var req UnlockUserReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid json", nil)
		return
	}

	h.users.UnlockUser(r.Context(), req.UserId)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("User unlocked")
}
