// Package client implements HTTP clients for marketplace APIs
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/config"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/domain"
)

const cardsBaseURL = "https://content-api.wildberries.ru"

// WbCardsClient is an HTTP client for the Wildberries Content API
type WbCardsClient struct {
	cfg    config.WbConfig
	client *retryablehttp.Client
}

// NewWbCardsClient returns a new WbCardsClient instance with configured retry policy
func NewWbCardsClient(cfg config.WbConfig) *WbCardsClient {
	rc := retryablehttp.NewClient()
	rc.Logger = nil
	rc.RetryMax = cfg.MaxRetries
	rc.RetryWaitMin = time.Duration(cfg.PaginationDelayMs) * time.Millisecond
	rc.RetryWaitMax = time.Duration(cfg.PaginationDelayMs) * time.Millisecond
	rc.CheckRetry = retryPolicy
	rc.HTTPClient.Timeout = cfg.Timeout

	return &WbCardsClient{
		cfg:    cfg,
		client: rc,
	}
}

type cardsRequest struct {
	Settings struct {
		Cursor *cardsCursor `json:"cursor,omitempty"`
		Filter struct {
			WithPhoto int `json:"withPhoto"`
		} `json:"filter"`
	} `json:"settings"`
	Limit int `json:"limit"`
}

type cardsCursor struct {
	UpdatedAt string `json:"updatedAt"`
	NmID      int64  `json:"nmID"`
}

type cardsResponse struct {
	Cards  []domain.WbCard `json:"cards"`
	Cursor *cardsCursor    `json:"cursor"`
}

// FetchCardsBatch requests a single page of product cards from the Content API
func (c *WbCardsClient) FetchCardsBatch(ctx context.Context, cursor *domain.SyncCursorState, limit int) ([]domain.WbCard, *domain.SyncCursorState, error) {
	reqURL := cardsBaseURL + "/content/v2/get/cards/list"

	body := cardsRequest{
		Limit: limit,
	}
	body.Settings.Filter.WithPhoto = -1

	if cursor != nil {
		body.Settings.Cursor = &cardsCursor{
			UpdatedAt: cursor.UpdatedAt,
			NmID:      cursor.NmID,
		}
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal cards request: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, reqURL, payload)
	if err != nil {
		return nil, nil, fmt.Errorf("create cards request: %w", err)
	}
	req.Header.Set("Authorization", c.cfg.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("execute cards request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read cards response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf("wb cards api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result cardsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, nil, fmt.Errorf("unmarshal cards response: %w", err)
	}

	var nextCursor *domain.SyncCursorState
	if result.Cursor != nil {
		nextCursor = &domain.SyncCursorState{
			UpdatedAt: result.Cursor.UpdatedAt,
			NmID:      result.Cursor.NmID,
		}
	}

	return result.Cards, nextCursor, nil
}

// FetchAllCards iterates over all cards pages, calling onBatch for each page.
// Returns the total number of cards fetched, the number of batches, and an error if any.
func (c *WbCardsClient) FetchAllCards(
	ctx context.Context,
	startCursor *domain.SyncCursorState,
	limit int,
	onBatch func(cards []domain.WbCard, cursor *domain.SyncCursorState) error,
) (int, int, error) {
	currentCursor := startCursor
	totalCards := 0
	batchesCount := 0

	for {
		batchesCount++
		cards, nextCursor, err := c.FetchCardsBatch(ctx, currentCursor, limit)
		if err != nil {
			return totalCards, batchesCount, fmt.Errorf("fetch cards batch %d: %w", batchesCount, err)
		}

		totalCards += len(cards)

		if len(cards) > 0 {
			if err := onBatch(cards, nextCursor); err != nil {
				return totalCards, batchesCount, err
			}
		}

		if nextCursor == nil || len(cards) < limit {
			break
		}
		currentCursor = nextCursor
	}

	return totalCards, batchesCount, nil
}
