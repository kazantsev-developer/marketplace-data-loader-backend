// Package main starts the HTTP server for marketplace data API
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/config"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/repository"
	httphandler "github.com/kazantsev-developer/marketplace-data-loader-backend/internal/transport/http"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := repository.NewPool(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("create db pool: %v", err)
	}
	defer pool.Close()

	ozonOrderRepo := repository.NewOzonOrderRepo(pool)
	ozonRemainRepo := repository.NewOzonRemainRepo(pool)

	wbOrderRepo := repository.NewWbOrderRepo(pool)
	wbRemainRepo := repository.NewWbRemainRepo(pool)
	wbCardRepo := repository.NewWbCardRepo(pool)

	msRepo := repository.NewMoyskladRepo(pool)

	logRepo := repository.NewSyncLogRepo(pool)

	handler := httphandler.NewHandler(ozonOrderRepo, ozonRemainRepo, wbOrderRepo, wbRemainRepo, wbCardRepo, msRepo, logRepo)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	go func() {
		log.Printf("server starting on :%d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down server...")
	srv.Shutdown(context.Background())
}
