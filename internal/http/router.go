package httpx

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/sebaactis/wallet-go-api/internal/health"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	// Middlewares base
	r.Use(chimw.RequestID) // añade un request ID único
	r.Use(chimw.RealIP)    // detecta IP real detrás de proxies
	r.Use(chimw.Recoverer) // evita que un panic tumbe el server
	r.Use(Logger())        // logs estructurados
	r.Use(JSONContentType())
	r.Use(Timeout(8 * time.Second))

	hh := health.New()
	r.Get("/health", hh.Liveness)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("pong"))
		})
	})

	return r
}
