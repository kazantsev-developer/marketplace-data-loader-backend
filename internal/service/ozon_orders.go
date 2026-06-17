// Package service implements business logic for data synchronization
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/client"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/config"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/domain"
)

// OzonOrdersService manages Ozon orders synchronization
type OzonOrdersService struct {
	repo    domain.OzonOrderRepository
	client  *client.OzonOrdersClient
	logRepo domain.SyncLogRepository
	cfg     config.OzonConfig
}

// NewOzonOrdersService returns a new OzonOrdersService instance
func NewOzonOrdersService(
	repo domain.OzonOrderRepository,
	client *client.OzonOrdersClient,
	logRepo domain.SyncLogRepository,
	cfg config.OzonConfig,
) *OzonOrdersService {
	return &OzonOrdersService{
		repo:    repo,
		client:  client,
		logRepo: logRepo,
		cfg:     cfg,
	}
}

// SyncOzonOrders performs the full Ozon orders synchronization for FBO and FBS
func (s *OzonOrdersService) SyncOzonOrders(ctx context.Context) error {
	startTime := time.Now()
	from, to := s.calculateDateRange()
	status := "success"
	var errMsg string
	var totalProcessed int

	log.Printf("[ozon-orders] sync started: %s — %s", from, to)

	defer func() {
		_, logErr := s.logRepo.InsertOzonLog(ctx, domain.OzonSyncLog{
			SyncAt:       startTime,
			Status:       status,
			RecordsCount: totalProcessed,
			DateFrom:     &startTime,
			ErrorMessage: stringPtr(errMsg),
		})
		if logErr != nil {
			log.Printf("[ozon-orders] failed to save log: %v", logErr)
		}
		log.Printf("[ozon-orders] sync finished: status=%s records=%d", status, totalProcessed)
	}()

	fboProcessed, err := s.syncScheme(ctx, "FBO", from, to)
	if err != nil {
		status = "error"
		errMsg = fmt.Sprintf("fbo sync: %v", err)
		log.Printf("[ozon-orders] FBO error: %v", err)
	}
	totalProcessed += fboProcessed

	fbsProcessed, err := s.syncScheme(ctx, "FBS", from, to)
	if err != nil {
		status = "error"
		errMsg = fmt.Sprintf("fbs sync: %v", err)
		log.Printf("[ozon-orders] FBS error: %v", err)
	}
	totalProcessed += fbsProcessed

	return nil
}

func (s *OzonOrdersService) syncScheme(ctx context.Context, scheme, from, to string) (int, error) {
	var processed int
	var err error

	if scheme == "FBO" {
		_, _, err = s.client.FetchAllFBOOrders(ctx, from, to, func(rawOrders []json.RawMessage, _ string) error {
			orders, normalizeErr := s.normalizeOrders(rawOrders, "FBO")
			if normalizeErr != nil {
				return normalizeErr
			}
			count, saveErr := s.repo.UpsertBatch(ctx, orders)
			if saveErr != nil {
				return fmt.Errorf("upsert fbo orders: %w", saveErr)
			}
			processed += count
			return nil
		})
	} else {
		_, _, err = s.client.FetchAllFBSOrders(ctx, from, to, func(rawOrders []json.RawMessage, _ string) error {
			orders, normalizeErr := s.normalizeOrders(rawOrders, "FBS")
			if normalizeErr != nil {
				return normalizeErr
			}
			count, saveErr := s.repo.UpsertBatch(ctx, orders)
			if saveErr != nil {
				return fmt.Errorf("upsert fbs orders: %w", saveErr)
			}
			processed += count
			return nil
		})
	}

	if err != nil {
		return processed, fmt.Errorf("fetch %s orders: %w", scheme, err)
	}
	return processed, nil
}

func (s *OzonOrdersService) normalizeOrders(rawOrders []json.RawMessage, scheme string) ([]domain.OzonOrder, error) {
	orders := make([]domain.OzonOrder, 0, len(rawOrders))
	for _, raw := range rawOrders {
		var o domain.OzonOrder
		if err := json.Unmarshal(raw, &o); err != nil {
			return nil, fmt.Errorf("unmarshal ozon order: %w", err)
		}
		o.Scheme = &scheme
		orders = append(orders, o)
	}
	return orders, nil
}

func (s *OzonOrdersService) calculateDateRange() (string, string) {
	now := time.Now().UTC()
	to := now.Format("2006-01-02T15:04:05") + "Z"
	from := now.AddDate(0, 0, -30).Format("2006-01-02T15:04:05") + "Z"
	return from, to
}
