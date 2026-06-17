// Package service implements business logic for data synchronization
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/client"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/domain"
)

// OzonStocksService manages Ozon stock synchronization
type OzonStocksService struct {
	repo    domain.OzonRemainRepository
	client  *client.OzonStocksClient
	logRepo domain.SyncLogRepository
}

// NewOzonStocksService returns a new OzonStocksService instance
func NewOzonStocksService(
	repo domain.OzonRemainRepository,
	client *client.OzonStocksClient,
	logRepo domain.SyncLogRepository,
) *OzonStocksService {
	return &OzonStocksService{
		repo:    repo,
		client:  client,
		logRepo: logRepo,
	}
}

// SyncOzonStocks performs the full Ozon stock synchronization flow
func (s *OzonStocksService) SyncOzonStocks(ctx context.Context) error {
	startTime := time.Now()
	status := "success"
	var errMsg string
	var totalProcessed int

	log.Println("[ozon-stocks] sync started")

	defer func() {
		_, logErr := s.logRepo.InsertOzonLog(ctx, domain.OzonSyncLog{
			SyncAt:       startTime,
			Status:       status,
			Scheme:       "stocks",
			RecordsCount: totalProcessed,
			DateFrom:     nil,
			ErrorMessage: stringPtr(errMsg),
		})
		if logErr != nil {
			log.Printf("[ozon-stocks] failed to save log: %v", logErr)
		}
		log.Printf("[ozon-stocks] sync finished: status=%s records=%d", status, totalProcessed)
	}()

	_, _, err := s.client.FetchAllStocks(ctx, func(rawStocks []json.RawMessage) error {
		remains, normalizeErr := normalizeStocks(rawStocks)
		if normalizeErr != nil {
			return fmt.Errorf("normalize stocks: %w", normalizeErr)
		}

		if len(remains) > 0 {
			count, upsertErr := s.repo.UpsertBatch(ctx, remains)
			if upsertErr != nil {
				return fmt.Errorf("upsert stocks: %w", upsertErr)
			}
			totalProcessed += count
		}
		return nil
	})
	if err != nil {
		status = "error"
		errMsg = fmt.Sprintf("fetch stocks: %v", err)
		log.Printf("[ozon-stocks] error: %v", err)
		return fmt.Errorf("fetch ozon stocks: %w", err)
	}

	_, err = s.repo.ResetStale(ctx, startTime.Format(time.RFC3339))
	if err != nil {
		status = "error"
		errMsg = fmt.Sprintf("reset stale stocks: %v", err)
		log.Printf("[ozon-stocks] error: %v", err)
		return fmt.Errorf("reset stale stocks: %w", err)
	}

	return nil
}

type rawStockItem struct {
	ProductID    int64  `json:"product_id"`
	OfferID      string `json:"offer_id"`
	HasFboStocks bool   `json:"has_fbo_stocks"`
}

func normalizeStocks(rawStocks []json.RawMessage) ([]domain.OzonRemain, error) {
	remains := make([]domain.OzonRemain, 0, len(rawStocks))
	for _, raw := range rawStocks {
		var item rawStockItem
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, fmt.Errorf("unmarshal stock item: %w", err)
		}
		remains = append(remains, domain.OzonRemain{
			Sku:              item.ProductID,
			ProductID:        item.ProductID,
			ItemCode:         stringPtr(item.OfferID),
			Name:             stringPtr(item.OfferID),
			FboVisibleAmount: boolToInt(item.HasFboStocks),
			FboPresentAmount: boolToInt(item.HasFboStocks),
		})
	}
	return remains, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
