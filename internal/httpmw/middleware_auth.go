package httpmw

import (
	"context"
	"net/http"

	"github.com/sebaactis/wallet-go-api/internal/auth"
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
	jwt         *auth.JWT
	userService *user.Service
}

func NewAuthMiddleware(jwt *auth.JWT, userService *user.Service) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt, userService: userService}
}

func (a *AuthMiddleware) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			accessCookie, err := r.Cookie("accessToken")
			
			if err == nil {
				userID, _, tokenType, parseErr := a.jwt.Parse(accessCookie.Value)
				
				if parseErr == nil && tokenType == auth.TokenTypeAccess {
					ctx := context.WithValue(r.Context(), ctxUserID, userID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			
			refreshCookie, err := r.Cookie("refreshToken")
			if err != nil {
				httputil.WriteError(w, http.StatusUnauthorized, "authentication required, please login", nil)
				return
			}
			
			refreshUserID, refreshEmail, refreshTokenType, err := a.jwt.Parse(refreshCookie.Value)
			if err != nil || refreshTokenType != auth.TokenTypeRefresh {
				httputil.WriteError(w, http.StatusUnauthorized, "session expired, please login again", nil)
				return
			}
			
			newAccessToken, err := a.jwt.Sign(refreshUserID, refreshEmail, auth.TokenTypeAccess)
			if err != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "failed to refresh token", nil)
				return
			}
			

			a.setAccessTokenCookie(w, auth.TokenTypeAccess, newAccessToken)
			
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
		MaxAge:   int(a.jwt.GetTTL(tokenType).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}
