package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sebaactis/wallet-go-api/internal/auth"
	"github.com/sebaactis/wallet-go-api/internal/entities/account"
	"github.com/sebaactis/wallet-go-api/internal/entities/token"
	"github.com/sebaactis/wallet-go-api/internal/entities/user"
	"github.com/sebaactis/wallet-go-api/internal/entities/wallet"
	httpx "github.com/sebaactis/wallet-go-api/internal/http"
	"github.com/sebaactis/wallet-go-api/internal/httpmw"
	"github.com/sebaactis/wallet-go-api/internal/platform/config"
	"github.com/sebaactis/wallet-go-api/internal/platform/database"
	"github.com/sebaactis/wallet-go-api/internal/validation"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("open db %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	validator := validation.NewValidator()
	rateLimiter := httpmw.NewRateLimiter(10, time.Minute*1)
	jwt := auth.NewJWT()

	// Repositorios
	userRepo := user.NewRepository(db)
	accountRepo := account.NewRepository(db)
	tokenRepo := token.NewRepository(db)

	// Servicios
	userService := user.NewService(userRepo, validator)
	accountService := account.NewService(accountRepo)
	walletService := wallet.NewService(db)
	tokenService := token.NewService(tokenRepo, validator)

	// Handlers

	userHandler := user.NewHTTPHandler(userService)
	accountHandler := account.NewHTTPHandler(accountService)
	walletHandler := wallet.NewHTTPHandler(walletService)
	authHandler := auth.NewHTTPHandler(userService, tokenService,jwt, validator)
	tokenHandler := token.NewHTTPHandler(tokenService)
	authMiddleware := httpmw.NewAuthMiddleware(jwt)

	r := httpx.NewRouter(
		httpx.Deps{
			UserHandler:    userHandler,
			AccountHandler: accountHandler,
			WalletHandler:  walletHandler,
			Validator:      validator,
			RateLimiter:    rateLimiter,
			AuthHandler:    authHandler,
			AuthMiddleWare: authMiddleware,
			TokensHandler: tokenHandler,
		},
	)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("API escuchando en %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}

	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	log.Println("Apago limpio")

}
