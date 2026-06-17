// Package main provides the entry point for marketplace synchronization jobs.
//
// Usage: go run ./cmd/sync --entity=wb_orders|wb_remains|wb_cards|ozon_orders|ozon_stocks|ms_stocks
package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/client"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/config"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/repository"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/service"
)

func main() {
	entity := flag.String("entity", "", "entity to sync: wb_orders, wb_remains, wb_cards, ozon_orders, ozon_stocks, ms_stocks")
	flag.Parse()

	if *entity == "" {
		log.Fatal("--entity flag is required")
	}

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

	syncLogRepo := repository.NewSyncLogRepo(pool)

	switch *entity {
	case "wb_orders":
		runWbOrders(ctx, cfg, pool, syncLogRepo)
	case "wb_remains":
		runWbRemains(ctx, cfg, pool, syncLogRepo)
	case "wb_cards":
		runWbCards(ctx, cfg, pool, syncLogRepo)
	case "ozon_orders":
		runOzonOrders(ctx, cfg, pool, syncLogRepo)
	case "ozon_stocks":
		runOzonStocks(ctx, cfg, pool, syncLogRepo)
	case "ms_stocks":
		runMsStocks(ctx, cfg, pool, syncLogRepo)
	default:
		log.Fatalf("unknown entity: %s", *entity)
	}
}

func runWbOrders(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, logRepo *repository.SyncLogRepo) {
	wbClient := client.NewWbClient(cfg.WB)
	orderRepo := repository.NewWbOrderRepo(pool)
	svc := service.NewOrdersService(orderRepo, wbClient, logRepo, cfg.WB, cfg.Settings.APILimit)
	if err := svc.SyncOrders(ctx); err != nil {
		log.Fatalf("wb orders sync failed: %v", err)
	}
}

func runWbRemains(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, logRepo *repository.SyncLogRepo) {
	remainsClient := client.NewWbRemainsClient(cfg.WB)
	remainsRepo := repository.NewWbRemainRepo(pool)
	svc := service.NewRemainsService(remainsRepo, remainsClient, logRepo)
	if err := svc.SyncRemains(ctx); err != nil {
		log.Fatalf("wb remains sync failed: %v", err)
	}
}

func runWbCards(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, logRepo *repository.SyncLogRepo) {
	cardsClient := client.NewWbCardsClient(cfg.WB)
	cardsRepo := repository.NewWbCardRepo(pool)
	svc := service.NewCardsService(cardsRepo, cardsClient, logRepo, cfg.Settings.BatchSize, cfg.Settings.BatchDelayMs)
	if err := svc.SyncCards(ctx); err != nil {
		log.Fatalf("wb cards sync failed: %v", err)
	}
}

func runOzonOrders(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, logRepo *repository.SyncLogRepo) {
	ozonClient := client.NewOzonOrdersClient(cfg.Ozon)
	orderRepo := repository.NewOzonOrderRepo(pool)
	svc := service.NewOzonOrdersService(orderRepo, ozonClient, logRepo, cfg.Ozon)
	if err := svc.SyncOzonOrders(ctx); err != nil {
		log.Fatalf("ozon orders sync failed: %v", err)
	}
}

func runOzonStocks(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, logRepo *repository.SyncLogRepo) {
	ozonStocksClient := client.NewOzonStocksClient(cfg.Ozon)
	stocksRepo := repository.NewOzonRemainRepo(pool)
	svc := service.NewOzonStocksService(stocksRepo, ozonStocksClient, logRepo)
	if err := svc.SyncOzonStocks(ctx); err != nil {
		log.Fatalf("ozon stocks sync failed: %v", err)
	}
}

func runMsStocks(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, logRepo *repository.SyncLogRepo) {
	msClient := client.NewMoyskladClient(cfg.MS)
	msRepo := repository.NewMoyskladRepo(pool)
	svc := service.NewMoyskladService(msRepo, msClient, logRepo, cfg.MS)
	if err := svc.SyncMoysklad(ctx); err != nil {
		log.Fatalf("ms stocks sync failed: %v", err)
	}
}
