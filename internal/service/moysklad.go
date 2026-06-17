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

// MoyskladService manages MoySklad stock synchronization
type MoyskladService struct {
	repo    domain.MoyskladRepository
	client  *client.MoyskladClient
	logRepo domain.SyncLogRepository
	cfg     config.MoyskladConfig
}

// NewMoyskladService returns a new MoyskladService instance
func NewMoyskladService(
	repo domain.MoyskladRepository,
	client *client.MoyskladClient,
	logRepo domain.SyncLogRepository,
	cfg config.MoyskladConfig,
) *MoyskladService {
	return &MoyskladService{
		repo:    repo,
		client:  client,
		logRepo: logRepo,
		cfg:     cfg,
	}
}

// SyncMoysklad performs the full MoySklad stock synchronization flow
func (s *MoyskladService) SyncMoysklad(ctx context.Context) error {
	startTime := time.Now()
	status := "success"
	var errMsg string
	var storesCount, stockRowsCount, detailsCount, aggregatesCount int

	log.Println("[moysklad] sync started")

	defer func() {
		execTime := int(time.Since(startTime).Seconds())
		_, logErr := s.logRepo.InsertMsJobLog(ctx, domain.MsJobLog{
			StartedAt:            startTime,
			Status:               status,
			RecordsCount:         stockRowsCount,
			DetailsCount:         detailsCount,
			AggregatesCount:      aggregatesCount,
			ErrorMessage:         stringPtr(errMsg),
			ExecutionTimeSeconds: &execTime,
		})
		if logErr != nil {
			log.Printf("[moysklad] failed to save job log: %v", logErr)
		}
		log.Printf("[moysklad] sync finished: status=%s stores=%d rows=%d details=%d aggregates=%d duration=%ds",
			status, storesCount, stockRowsCount, detailsCount, aggregatesCount, execTime)
	}()

	rawStores, err := s.client.FetchStores(ctx)
	if err != nil {
		status = "error"
		errMsg = fmt.Sprintf("fetch stores: %v", err)
		return fmt.Errorf("fetch stores: %w", err)
	}

	stores, err := parseStores(rawStores)
	if err != nil {
		status = "error"
		errMsg = fmt.Sprintf("parse stores: %v", err)
		return fmt.Errorf("parse stores: %w", err)
	}

	if len(stores) > 0 {
		count, err := s.repo.UpsertStores(ctx, stores)
		if err != nil {
			status = "error"
			errMsg = fmt.Sprintf("upsert stores: %v", err)
			return fmt.Errorf("upsert stores: %w", err)
		}
		storesCount = count
		log.Printf("[moysklad] saved %d stores", storesCount)
	}

	snapshotID, err := s.repo.CreateSnapshot(ctx)
	if err != nil {
		status = "error"
		errMsg = fmt.Sprintf("create snapshot: %v", err)
		return fmt.Errorf("create snapshot: %w", err)
	}
	log.Printf("[moysklad] snapshot created: %d", snapshotID)

	rawRows, err := s.client.FetchAllStockByStore(ctx)
	if err != nil {
		status = "error"
		errMsg = fmt.Sprintf("fetch stock: %v", err)
		return fmt.Errorf("fetch stock: %w", err)
	}
	stockRowsCount = len(rawRows)
	log.Printf("[moysklad] fetched %d stock rows", stockRowsCount)

	if stockRowsCount == 0 {
		return nil
	}

	details, aggregates, err := normalizeStockData(rawRows, snapshotID)
	if err != nil {
		status = "error"
		errMsg = fmt.Sprintf("normalize stock: %v", err)
		return fmt.Errorf("normalize stock: %w", err)
	}
	detailsCount = len(details)
	aggregatesCount = len(aggregates)

	log.Printf("[moysklad] normalized: %d details, %d aggregates", detailsCount, aggregatesCount)

	if len(details) > 0 {
		count, err := s.repo.InsertStockDetails(ctx, details)
		if err != nil {
			status = "error"
			errMsg = fmt.Sprintf("insert stock details: %v", err)
			return fmt.Errorf("insert stock details: %w", err)
		}
		log.Printf("[moysklad] saved %d stock details", count)
	}

	if len(aggregates) > 0 {
		count, err := s.repo.UpsertProductTotals(ctx, aggregates)
		if err != nil {
			status = "error"
			errMsg = fmt.Sprintf("upsert product totals: %v", err)
			return fmt.Errorf("upsert product totals: %w", err)
		}
		log.Printf("[moysklad] saved %d product aggregates", count)
	}

	return nil
}

type rawStore struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	ExternalCode string `json:"externalCode"`
	Address      string `json:"address"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
}

func parseStores(rawStores []json.RawMessage) ([]domain.MsStore, error) {
	stores := make([]domain.MsStore, 0, len(rawStores))
	for _, raw := range rawStores {
		var s rawStore
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, fmt.Errorf("unmarshal store: %w", err)
		}
		createdAt, _ := time.Parse(time.RFC3339, s.Created)
		updatedAt, _ := time.Parse(time.RFC3339, s.Updated)
		stores = append(stores, domain.MsStore{
			UUID:         s.ID,
			Name:         s.Name,
			Code:         stringPtr(s.Code),
			ExternalCode: stringPtr(s.ExternalCode),
			Address:      stringPtr(s.Address),
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			SyncedAt:     time.Now(),
		})
	}
	return stores, nil
}

type rawStockRow struct {
	Meta         *rawMeta          `json:"meta"`
	Name         string            `json:"name"`
	Article      string            `json:"article"`
	StockByStore []rawStockByStore `json:"stockByStore"`
}

type rawMeta struct {
	Href string `json:"href"`
}

type rawStockByStore struct {
	Meta      *rawMeta `json:"meta"`
	Stock     float64  `json:"stock"`
	Reserve   float64  `json:"reserve"`
	InTransit float64  `json:"inTransit"`
}

func extractUUIDFromHref(href string) string {
	if href == "" {
		return ""
	}
	for i := len(href) - 1; i >= 0; i-- {
		if href[i] == '/' {
			return href[i+1:]
		}
	}
	return href
}

func normalizeStockData(rawRows []json.RawMessage, snapshotID int) ([]domain.MsStockDetail, []domain.MsProductTotal, error) {
	detailsMap := make(map[string]*domain.MsStockDetail)
	aggregatesMap := make(map[string]*domain.MsProductTotal)

	for _, raw := range rawRows {
		var row rawStockRow
		if err := json.Unmarshal(raw, &row); err != nil {
			return nil, nil, fmt.Errorf("unmarshal stock row: %w", err)
		}

		productUUID := ""
		if row.Meta != nil {
			productUUID = extractUUIDFromHref(row.Meta.Href)
		}
		if productUUID == "" {
			continue
		}

		if _, exists := aggregatesMap[productUUID]; !exists {
			aggregatesMap[productUUID] = &domain.MsProductTotal{
				ProductUUID: productUUID,
				Article:     stringPtr(row.Article),
				Name:        stringPtr(row.Name),
				SnapshotID:  &snapshotID,
			}
		}
		totals := aggregatesMap[productUUID]

		for _, sbs := range row.StockByStore {
			storeUUID := ""
			if sbs.Meta != nil {
				storeUUID = extractUUIDFromHref(sbs.Meta.Href)
			}
			if storeUUID == "" {
				continue
			}

			stock := int(sbs.Stock)
			reserve := int(sbs.Reserve)
			inTransit := int(sbs.InTransit)

			if stock == 0 && reserve == 0 && inTransit == 0 {
				continue
			}

			key := productUUID + "_" + storeUUID
			if _, exists := detailsMap[key]; !exists {
				detailsMap[key] = &domain.MsStockDetail{
					SnapshotID:  snapshotID,
					ProductUUID: productUUID,
					StoreUUID:   storeUUID,
				}
			}
			detail := detailsMap[key]
			detail.Stock += stock
			detail.Reserve += reserve
			detail.InTransit += inTransit

			totals.TotalStock += stock
			totals.TotalReserve += reserve
			totals.TotalInTransit += inTransit
		}
	}

	details := make([]domain.MsStockDetail, 0, len(detailsMap))
	for _, d := range detailsMap {
		details = append(details, *d)
	}
	aggregates := make([]domain.MsProductTotal, 0, len(aggregatesMap))
	for _, a := range aggregatesMap {
		aggregates = append(aggregates, *a)
	}

	return details, aggregates, nil
}
