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
	httpx "github.com/sebaactis/wallet-go-api/internal/http"
	"github.com/sebaactis/wallet-go-api/internal/platform/config"
	"github.com/sebaactis/wallet-go-api/internal/platform/database"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	db, err := database.Open(cfg); if err != nil {
		log.Fatalf("open db %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	r := httpx.NewRouter()

	srv := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: r,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout: 60 * time.Second,
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

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	log.Println("Apago limpio")


}

