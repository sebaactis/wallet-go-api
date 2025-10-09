package httpmw

import (
	"context"
	"net/http"
	"strings"

	"github.com/sebaactis/wallet-go-api/internal/auth"
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
	jwt *auth.JWT
}

func NewAuthMiddleware(jwt *auth.JWT) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
}

func (a *AuthMiddleware) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
				httputil.WriteError(w, http.StatusUnauthorized, "missing bearer token", nil)
				return
			}
			tok := strings.TrimSpace(h[len("Bearer "):])
			userID, _, err := a.jwt.Parse(tok)
			if err != nil {
				httputil.WriteError(w, http.StatusUnauthorized, "invalid token", nil)
				return
			}
			ctx := context.WithValue(r.Context(), ctxUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
