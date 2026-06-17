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

// OrdersService manages Wildberries orders synchronization
type OrdersService struct {
	repo     domain.WbOrderRepository
	client   *client.WbClient
	logRepo  domain.SyncLogRepository
	cfg      config.WbConfig
	apiLimit int
}

// NewOrdersService returns a new OrdersService instance
func NewOrdersService(
	repo domain.WbOrderRepository,
	client *client.WbClient,
	logRepo domain.SyncLogRepository,
	cfg config.WbConfig,
	apiLimit int,
) *OrdersService {
	return &OrdersService{
		repo:     repo,
		client:   client,
		logRepo:  logRepo,
		cfg:      cfg,
		apiLimit: apiLimit,
	}
}

// SyncOrders performs the full orders synchronization flow
func (s *OrdersService) SyncOrders(ctx context.Context) error {
	startTime := time.Now()
	from, to := calculateDateRange()
	var totalProcessed int
	status := "success"
	var errMsg string

	log.Printf("[orders] sync started: %s — %s", from.Format(time.RFC3339), to.Format(time.RFC3339))

	defer func() {
		execTime := int(time.Since(startTime).Seconds())
		_, logErr := s.logRepo.Insert(ctx, domain.SyncLog{
			SyncAt:               startTime,
			Status:               status,
			RecordsCount:         totalProcessed,
			ExecutionTimeSeconds: execTime,
			EntityType:           "wb_orders",
			ErrorMessage:         stringPtr(errMsg),
		})
		if logErr != nil {
			log.Printf("[orders] failed to save log: %v", logErr)
		}
		log.Printf("[orders] sync finished: status=%s records=%d duration=%ds", status, totalProcessed, execTime)
	}()

	currentDateFrom := from.Format(time.RFC3339)
	var pageCount int

	for {
		pageCount++
		log.Printf("[orders] fetching page %d, dateFrom=%s", pageCount, currentDateFrom)

		rawOrders, err := s.client.FetchOrders(ctx, currentDateFrom)
		if err != nil {
			status = "error"
			errMsg = fmt.Sprintf("fetch orders page %d: %v", pageCount, err)
			return fmt.Errorf("fetch orders: %w", err)
		}
		log.Printf("[orders] page %d: received %d orders", pageCount, len(rawOrders))

		orders, err := parseOrders(rawOrders)
		if err != nil {
			status = "error"
			errMsg = fmt.Sprintf("parse orders page %d: %v", pageCount, err)
			return fmt.Errorf("parse orders: %w", err)
		}

		filtered := filterOrdersByDate(orders, from, to)
		log.Printf("[orders] page %d: %d orders after date filter", pageCount, len(filtered))

		if len(filtered) > 0 {
			count, err := s.repo.UpsertBatch(ctx, filtered)
			if err != nil {
				status = "error"
				errMsg = fmt.Sprintf("upsert batch page %d: %v", pageCount, err)
				return fmt.Errorf("upsert orders batch: %w", err)
			}
			totalProcessed += count
			log.Printf("[orders] page %d: saved %d orders", pageCount, count)
		}

		if len(rawOrders) < s.apiLimit {
			log.Printf("[orders] last page reached (page size %d < api limit %d)", len(rawOrders), s.apiLimit)
			break
		}

		lastOrder := orders[len(orders)-1]
		currentDateFrom = lastOrder.LastChangeDate.Add(1 * time.Millisecond).Format(time.RFC3339)

		log.Printf("[orders] waiting %v (rate limit)...", time.Duration(s.cfg.PaginationDelayMs)*time.Millisecond)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(s.cfg.PaginationDelayMs) * time.Millisecond):
		}
	}

	return nil
}

func calculateDateRange() (time.Time, time.Time) {
	now := time.Now().UTC()
	to := now.AddDate(0, 0, -1)
	to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())
	from := now.AddDate(0, 0, -30)
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	return from, to
}

func parseOrders(raw []json.RawMessage) ([]domain.WbOrder, error) {
	orders := make([]domain.WbOrder, 0, len(raw))
	for i, r := range raw {
		var o domain.WbOrder
		if err := json.Unmarshal(r, &o); err != nil {
			return nil, fmt.Errorf("unmarshal order %d: %w", i, err)
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func filterOrdersByDate(orders []domain.WbOrder, from, to time.Time) []domain.WbOrder {
	filtered := orders[:0]
	for _, o := range orders {
		if !o.Date.Before(from) && !o.Date.After(to) {
			filtered = append(filtered, o)
		}
	}
	return filtered
}
