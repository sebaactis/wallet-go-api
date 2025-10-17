package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sebaactis/wallet-go-api/internal/entities/token"
	"github.com/sebaactis/wallet-go-api/internal/entities/user"
	"github.com/sebaactis/wallet-go-api/internal/httputil"
	"github.com/sebaactis/wallet-go-api/internal/validation"
	"golang.org/x/crypto/bcrypt"
)

type HTTPHandler struct {
	users     *user.Service
	tokens    *token.Service
	jwt       *JWT
	validator validation.StructValidator
}

func NewHTTPHandler(users *user.Service, tokens *token.Service, jwt *JWT, validator validation.StructValidator) *HTTPHandler {
	return &HTTPHandler{users: users, tokens: tokens, jwt: jwt, validator: validator}
}

func (h *HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseLoginRequest(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	user, err := h.authenticateUser(r.Context(), req)
	if err != nil {
		h.handleLoginError(w, r.Context(), err, user)
		return
	}

	tokens, err := h.generateTokens(r.Context(), user)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "cannot generate tokens", nil)
		return
	}
	h.setTokenCookie(w, "accessToken", tokens.AccessToken, TokenTypeAccess)
	h.setTokenCookie(w, "refreshToken", tokens.RefreshToken, TokenTypeRefresh)
	h.respondWithTokens(w, user, tokens)
}

func (h *HTTPHandler) RecoveryPasswordRequest(w http.ResponseWriter, r *http.Request) {
	var req RecoveryPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid json", nil)
		return
	}

	user, err := h.users.GetByEmail(r.Context(), req.Email)

	if err != nil {
		httputil.WriteError(w, http.StatusOK, "if the mail exists, a recovery link will be sent you", nil)
		return
	}

	token, err := h.generateTokenRecovery(r.Context(), user)

	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Token generate failed", nil)
		return
	}

	response := &RecoveryPasswordRequestResponse{
		Email: req.Email,
		Token: *token,
	}

	httputil.WriteJSON(w, http.StatusOK, response)
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

func (h *HTTPHandler) UpdatePasswordByRecovery(w http.ResponseWriter, r *http.Request) {
	var req user.UserRecoveryPassword

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
		return
	}

	if _, _, _, err := h.jwt.ParseResetPassword(req.Token); err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	userRecovery, err := h.users.UpdatePasswordByRecovery(r.Context(), (req))

	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, user.ToResponse(userRecovery))

}




// ==================== MÉTODOS PRIVADOS ====================

func (h *HTTPHandler) parseLoginRequest(r *http.Request) (*LoginRequest, error) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.New("invalid json")
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if fields, ok := h.validator.ValidateStruct(&req); !ok {
		return nil, fmt.Errorf("validation error: %v", fields)
	}

	return &req, nil
}

func (h *HTTPHandler) authenticateUser(ctx context.Context, req *LoginRequest) (*user.User, error) {

	user, err := h.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Locked_until.After(time.Now()) {
		return user, ErrAccountLocked
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return user, ErrInvalidCredentials
	}

	return user, nil
}

func (h *HTTPHandler) generateTokens(ctx context.Context, user *user.User) (*TokenPair, error) {
	accessToken, err := h.jwt.Sign(user.ID, user.Email, TokenTypeAccess)
	if err != nil {
		return nil, err
	}

	refreshToken, err := h.jwt.Sign(user.ID, user.Email, TokenTypeRefresh)
	if err != nil {
		return nil, err
	}

	if _, err = h.tokens.Create(ctx, &token.TokenRequest{
		TokenType: string(TokenTypeAccess),
		Token:     accessToken,
	}); err != nil {
		return nil, err
	}

	if _, err = h.tokens.Create(ctx, &token.TokenRequest{
		TokenType: string(TokenTypeRefresh),
		Token:     refreshToken,
	}); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *HTTPHandler) generateTokenRecovery(ctx context.Context, user *user.User) (*string, error) {

	recoveryToken, err := h.jwt.Sign(user.ID, user.Email, TokenTypeResetPassword)
	if err != nil {
		return nil, err
	}

	if _, err = h.tokens.Create(ctx, &token.TokenRequest{
		TokenType: string(TokenTypeResetPassword),
		Token:     recoveryToken,
	}); err != nil {
		return nil, err
	}

	return &recoveryToken, nil
}

func (h *HTTPHandler) handleLoginError(w http.ResponseWriter, ctx context.Context, err error, user *user.User) {
	switch err {
	case ErrAccountLocked:
		httputil.WriteError(w, http.StatusLocked, "account temporarily locked", nil)

	case ErrInvalidCredentials:
		httputil.WriteError(w, http.StatusUnauthorized, "invalid credentials", nil)

		// Incrementar intentos solo si el usuario existe
		if user != nil {
			h.users.IncrementLoginAttempt(ctx, user.ID)
		}

	default:
		httputil.WriteError(w, http.StatusInternalServerError, "internal error", nil)
	}
}

func (h *HTTPHandler) setTokenCookie(w http.ResponseWriter, name string, token string, tokenType TokenType) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		MaxAge:   int(h.jwt.GetTTL(tokenType).Seconds() * 2),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *HTTPHandler) respondWithTokens(w http.ResponseWriter, user *user.User, tokens *TokenPair) {
	response := LoginResponse{
		Email:        user.Email,
		Name:         user.Name,
		Token:        tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account locked")
)
