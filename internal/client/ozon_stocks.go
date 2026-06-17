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
)

// OzonStocksClient is an HTTP client for the Ozon Product API (stock levels)
type OzonStocksClient struct {
	cfg    config.OzonConfig
	client *retryablehttp.Client
}

// NewOzonStocksClient returns a new OzonStocksClient instance with configured retry policy
func NewOzonStocksClient(cfg config.OzonConfig) *OzonStocksClient {
	rc := retryablehttp.NewClient()
	rc.Logger = nil
	rc.RetryMax = 3
	rc.RetryWaitMin = 200 * time.Millisecond
	rc.RetryWaitMax = 2 * time.Second
	rc.CheckRetry = retryPolicy
	rc.HTTPClient.Timeout = cfg.Timeout

	return &OzonStocksClient{
		cfg:    cfg,
		client: rc,
	}
}

type stocksRequest struct {
	Filter stocksFilter `json:"filter"`
	Limit  int          `json:"limit"`
	LastID string       `json:"last_id"`
}

type stocksFilter struct {
	Visibility string `json:"visibility"`
}

type stocksResponse struct {
	Result struct {
		Items  []json.RawMessage `json:"items"`
		LastID string            `json:"last_id"`
	} `json:"result"`
}

// FetchStocksBatch retrieves a single page of product stock data
func (c *OzonStocksClient) FetchStocksBatch(ctx context.Context, lastID string, limit int) ([]json.RawMessage, string, error) {
	reqBody := stocksRequest{
		Filter: stocksFilter{Visibility: "ALL"},
		Limit:  limit,
		LastID: lastID,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", fmt.Errorf("marshal stocks request: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/v3/product/list", bodyBytes)
	if err != nil {
		return nil, "", fmt.Errorf("create stocks request: %w", err)
	}
	req.Header.Set("Client-Id", c.cfg.ClientID)
	req.Header.Set("Api-Key", c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("execute stocks request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		limitedBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, "", fmt.Errorf("ozon stocks api returned status %d: %s", resp.StatusCode, string(limitedBody))
	}

	var result stocksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("unmarshal stocks response: %w", err)
	}

	return result.Result.Items, result.Result.LastID, nil
}

// FetchAllStocks fetches all product stock data, calling onBatch for each page
func (c *OzonStocksClient) FetchAllStocks(ctx context.Context, onBatch func(stocks []json.RawMessage) error) (int, int, error) {
	limit := 100
	lastID := ""
	totalProcessed := 0
	batchesCount := 0

	for {
		if err := ctx.Err(); err != nil {
			return totalProcessed, batchesCount, fmt.Errorf("context cancelled before batch: %w", err)
		}

		items, newLastID, err := c.FetchStocksBatch(ctx, lastID, limit)
		if err != nil {
			return totalProcessed, batchesCount, fmt.Errorf("fetch stocks batch: %w", err)
		}

		batchesCount++
		totalProcessed += len(items)

		if len(items) > 0 {
			if err := onBatch(items); err != nil {
				return totalProcessed, batchesCount, fmt.Errorf("process batch callback: %w", err)
			}
		}

		if newLastID == "" {
			break
		}
		lastID = newLastID

		if c.cfg.PaginationDelayMs > 0 {
			timer := time.NewTimer(time.Duration(c.cfg.PaginationDelayMs) * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return totalProcessed, batchesCount, fmt.Errorf("context cancelled during pagination delay: %w", ctx.Err())
			case <-timer.C:
			}
		}
	}

	return totalProcessed, batchesCount, nil
}
