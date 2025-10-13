package httpx

import (
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/sebaactis/wallet-go-api/internal/auth"
	"github.com/sebaactis/wallet-go-api/internal/entities/account"
	"github.com/sebaactis/wallet-go-api/internal/entities/token"
	"github.com/sebaactis/wallet-go-api/internal/entities/user"
	"github.com/sebaactis/wallet-go-api/internal/entities/wallet"
	"github.com/sebaactis/wallet-go-api/internal/health"
	"github.com/sebaactis/wallet-go-api/internal/httpmw"
	"github.com/sebaactis/wallet-go-api/internal/validation"
)

type Deps struct {
	UserHandler    *user.HTTPHandler
	AccountHandler *account.HTTPHandler
	WalletHandler  *wallet.HTTPHandler
	Validator      *validation.Validator
	RateLimiter    *httpmw.RateLimiter
	AuthHandler    *auth.HTTPHandler
	AuthMiddleWare *httpmw.AuthMiddleware
	TokensHandler *token.HTTPHandler
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
		r.Get("/users", d.UserHandler.FindAll)
		r.Post("/register", d.UserHandler.Create)
		r.Post("/login", d.AuthHandler.Login)
		r.Post("/unlock", d.AuthHandler.UnlockUser)
		r.Get("/tokens", d.TokensHandler.GetAll)

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
