package httpx

import (
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/sebaactis/wallet-go-api/internal/account"
	"github.com/sebaactis/wallet-go-api/internal/auth"
	"github.com/sebaactis/wallet-go-api/internal/health"
	"github.com/sebaactis/wallet-go-api/internal/httpmw"
	"github.com/sebaactis/wallet-go-api/internal/user"
	"github.com/sebaactis/wallet-go-api/internal/validation"
	"github.com/sebaactis/wallet-go-api/internal/wallet"
)

type Deps struct {
	UserHandler    *user.HTTPHandler
	AccountHandler *account.HTTPHandler
	WalletHandler  *wallet.HTTPHandler
	Validator      *validation.Validator
	RateLimiter    *httpmw.RateLimiter
	AuthHandler    *auth.HTTPHandler
	AuthMiddleWare *httpmw.AuthMiddleware
}

func NewRouter(d Deps) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.RequestID, chimw.RealIP, chimw.Recoverer)
	r.Use(httpmw.Logger(), httpmw.JSONContentType(), httpmw.Timeout(8*time.Second))

	if d.RateLimiter != nil {
		r.Use(d.RateLimiter.Middleware())
	}

	hh := health.New()
	r.Get("/health", hh.Liveness)

	r.Route("/v1", func(r chi.Router) {
		r.Post("/users", d.UserHandler.Create)
		r.Post("/auth/token", d.AuthHandler.Token)

		// Rutas protegidas:
		r.Group(func(pr chi.Router) {
			pr.Use(d.AuthMiddleWare.RequireAuth())

			pr.Get("/users/{id}", d.UserHandler.GetByID)
			pr.Post("/accounts", d.AccountHandler.Create)
			pr.Get("/accounts/{id}/balance", d.AccountHandler.GetBalance)

			pr.Post("/wallet/deposit", d.WalletHandler.Deposit)
			pr.Post("/wallet/withdraw", d.WalletHandler.Withdraw)
			pr.Post("/wallet/transfer", d.WalletHandler.Transfer)
		})
	})

	return r
}
