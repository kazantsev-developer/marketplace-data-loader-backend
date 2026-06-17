// Package service implements business logic for data synchronization
package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/client"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/domain"
)

// CardsService manages Wildberries product cards synchronization
type CardsService struct {
	repo       domain.WbCardRepository
	client     *client.WbCardsClient
	logRepo    domain.SyncLogRepository
	batchLimit int
	delayMs    int
}

// NewCardsService returns a new CardsService instance
func NewCardsService(
	repo domain.WbCardRepository,
	client *client.WbCardsClient,
	logRepo domain.SyncLogRepository,
	batchLimit int,
	delayMs int,
) *CardsService {
	if batchLimit <= 0 {
		batchLimit = 100
	}
	return &CardsService{
		repo:       repo,
		client:     client,
		logRepo:    logRepo,
		batchLimit: batchLimit,
		delayMs:    delayMs,
	}
}

// SyncCards performs the full cards synchronization flow (incremental if cursor exists)
func (s *CardsService) SyncCards(ctx context.Context) error {
	startTime := time.Now()
	status := "success"
	var errMsg string
	var totalProcessed int

	log.Println("[cards] sync started")

	defer func() {
		execTime := int(time.Since(startTime).Seconds())
		_, logErr := s.logRepo.Insert(ctx, domain.SyncLog{
			SyncAt:               startTime,
			Status:               status,
			RecordsCount:         totalProcessed,
			ExecutionTimeSeconds: execTime,
			EntityType:           "wb_cards",
			ErrorMessage:         stringPtr(errMsg),
		})
		if logErr != nil {
			log.Printf("[cards] failed to save log: %v", logErr)
		}
		log.Printf("[cards] sync finished: status=%s records=%d duration=%ds", status, totalProcessed, execTime)
	}()

	cursor, err := s.repo.GetCursor(ctx)
	if err != nil {
		status = "error"
		errMsg = fmt.Sprintf("get cursor: %v", err)
		return fmt.Errorf("get cursor: %w", err)
	}
	if cursor != nil {
		log.Printf("[cards] incremental sync from cursor updatedAt=%s nmID=%d", cursor.UpdatedAt, cursor.NmID)
	} else {
		log.Println("[cards] full sync (no cursor)")
	}

	var lastCursor *domain.SyncCursorState
	onBatch := func(cards []domain.WbCard, cur *domain.SyncCursorState) error {
		if len(cards) == 0 {
			return nil
		}
		count, err := s.repo.UpsertBatch(ctx, cards)
		if err != nil {
			return fmt.Errorf("upsert cards batch: %w", err)
		}
		totalProcessed += count
		log.Printf("[cards] batch saved: %d records", count)

		if cur != nil {
			if err := s.repo.SaveCursor(ctx, cur.UpdatedAt, cur.NmID); err != nil {
				return fmt.Errorf("save cursor: %w", err)
			}
			lastCursor = cur
		}
		if s.delayMs > 0 {
			time.Sleep(time.Duration(s.delayMs) * time.Millisecond)
		}
		return nil
	}

	totalFetched, batchesCount, fetchErr := s.client.FetchAllCards(ctx, cursor, s.batchLimit, onBatch)
	if fetchErr != nil {
		status = "error"
		errMsg = fmt.Sprintf("fetch cards: %v", fetchErr)
		return fetchErr
	}

	log.Printf("[cards] sync complete: fetched=%d batches=%d processed=%d", totalFetched, batchesCount, totalProcessed)

	if lastCursor != nil {
		log.Printf("[cards] final cursor: updatedAt=%s nmID=%d", lastCursor.UpdatedAt, lastCursor.NmID)
	}

	return nil
}
