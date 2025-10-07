package httpx

import (
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/sebaactis/wallet-go-api/internal/account"
	"github.com/sebaactis/wallet-go-api/internal/health"
	"github.com/sebaactis/wallet-go-api/internal/user"
)

type Deps struct {
	UserHandler    *user.HTTPHandler
	AccountHandler *account.HTTPHandler
}

func NewRouter(d Deps) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.RequestID, chimw.RealIP, chimw.Recoverer)
	r.Use(Logger(), JSONContentType(), Timeout(8*time.Second))

	hh := health.New()
	r.Get("/health", hh.Liveness)

	r.Route("/v1", func(r chi.Router) {
		r.Post("/users", d.UserHandler.Create)
		r.Get("/users/{id}", d.UserHandler.GetByID)

		r.Post("/accounts", d.AccountHandler.Create)
		r.Get("/accounts/{id}/balance", d.AccountHandler.GetBalance)
	})

	return r
}
