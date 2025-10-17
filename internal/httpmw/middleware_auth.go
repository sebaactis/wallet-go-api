package httpmw

import (
	"context"
	"net/http"

	"github.com/sebaactis/wallet-go-api/internal/auth"
	"github.com/sebaactis/wallet-go-api/internal/entities/token"
	"github.com/sebaactis/wallet-go-api/internal/entities/user"
	"github.com/sebaactis/wallet-go-api/internal/httputil"
)

type ctxKey string

const ctxUserID ctxKey = "auth.user_id"

func UserIDFromContext(ctx context.Context) (uint, bool) {
	v := ctx.Value(ctxUserID)
	if v == nil {
		return 0, false
	}

	id, ok := v.(uint)
	return id, ok
}

type AuthMiddleware struct {
	jwt          *auth.JWT
	userService  *user.Service
	tokenService *token.Service
}

func NewAuthMiddleware(jwt *auth.JWT, userService *user.Service, tokenService *token.Service) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt, userService: userService, tokenService: tokenService}
}

func (a *AuthMiddleware) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accessCookie, err := r.Cookie("accessToken")

			if err == nil {
				userID, _, tokenType, parseErr := a.jwt.Parse(accessCookie.Value, auth.TokenTypeAccess)

				if parseErr == nil && tokenType == auth.TokenTypeAccess {
					ctx := context.WithValue(r.Context(), ctxUserID, userID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}

				// Si el access token es inválido, revocarlo
				_ = a.tokenService.RevokeToken(r.Context(), accessCookie.Value)
				// Eliminar la cookie inmediatamente
				a.clearCookie(w, "accessToken")
			}

			refreshCookie, err := r.Cookie("refreshToken")
			if err != nil {
				httputil.WriteError(w, http.StatusUnauthorized, "authentication required, please login", nil)
				return
			}

			refreshUserID, refreshEmail, refreshTokenType, err := a.jwt.Parse(refreshCookie.Value, auth.TokenTypeRefresh)
			if err != nil || refreshTokenType != auth.TokenTypeRefresh {
				_ = a.tokenService.RevokeToken(r.Context(), refreshCookie.Value)
				a.clearCookie(w, "refreshToken")

				httputil.WriteError(w, http.StatusUnauthorized, "session expired, please login again", nil)
				return
			}

			newAccessToken, err := a.jwt.Sign(refreshUserID, refreshEmail, auth.TokenTypeAccess)
			if err != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "failed to refresh token", nil)
				return
			}

			a.setAccessTokenCookie(w, auth.TokenTypeAccess, newAccessToken)
			a.tokenService.Create(r.Context(), &token.TokenRequest{
				TokenType: string(auth.TokenTypeAccess),
				Token: newAccessToken,
			})

			ctx := context.WithValue(r.Context(), ctxUserID, refreshUserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *AuthMiddleware) setAccessTokenCookie(w http.ResponseWriter, tokenType auth.TokenType, accessToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(a.jwt.GetTTL(tokenType).Seconds() * 2),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (a *AuthMiddleware) clearCookie(w http.ResponseWriter, cookieName string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}
