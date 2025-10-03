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
	"github.com/sebaactis/wallet-go-api/internal/account"
	httpx "github.com/sebaactis/wallet-go-api/internal/http"
	"github.com/sebaactis/wallet-go-api/internal/platform/config"
	"github.com/sebaactis/wallet-go-api/internal/platform/database"
	"github.com/sebaactis/wallet-go-api/internal/user"
	"gorm.io/gorm"
)

func probe(ctx context.Context, db *gorm.DB) {

	userRepo := user.NewRepository(db)
	accountRepo := account.NewRepository(db)
	userService := user.NewService(userRepo)
	aService := account.NewService(accountRepo)

	userCreate := &user.UserCreate{
		Name:  "Sebastian Actis",
		Email: "sebaactis@gmail.com",
	}

	u, err := userService.Create(ctx, userCreate)
	if err != nil {
		log.Println("create user:", err)
		return
	}
	log.Println("user id:", u.ID)

	accountCreate := &account.AccountCreate{
		UserID:   u.ID,
		Currency: "ARS",
	}

	account, err := aService.Create(ctx, accountCreate)
	if err != nil {
		log.Println("create account:", err)
		return
	}
	log.Println("account id:", account.ID, "currency:", account.Currency)

	bal, err := aService.GetBalance(ctx, account.ID)
	if err != nil {
		log.Println("get balance:", err)
		return
	}
	log.Println("balance:", bal)
}

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

	r := httpx.NewRouter()

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

	probe(context.Background(), db)

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
