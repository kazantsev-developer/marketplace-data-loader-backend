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

// OzonOrdersClient is an HTTP client for the Ozon Seller API (FBO and FBS postings)
type OzonOrdersClient struct {
	cfg    config.OzonConfig
	client *retryablehttp.Client
}

// NewOzonOrdersClient returns a new OzonOrdersClient instance with configured retry policy
func NewOzonOrdersClient(cfg config.OzonConfig) *OzonOrdersClient {
	rc := retryablehttp.NewClient()
	rc.Logger = nil
	rc.RetryMax = 3
	rc.RetryWaitMin = 200 * time.Millisecond
	rc.RetryWaitMax = 2 * time.Second
	rc.CheckRetry = retryPolicy
	rc.HTTPClient.Timeout = cfg.Timeout

	return &OzonOrdersClient{
		cfg:    cfg,
		client: rc,
	}
}

type fboRequest struct {
	Dir    string      `json:"dir"`
	Filter fboFilter   `json:"filter"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
	With   fboWithOpts `json:"with"`
}

type fboFilter struct {
	Since string `json:"since"`
	To    string `json:"to"`
}

type fboWithOpts struct {
	AnalyticsData bool `json:"analytics_data"`
	FinancialData bool `json:"financial_data"`
}

type fboResponse struct {
	Result struct {
		Postings []json.RawMessage `json:"postings"`
		Total    int               `json:"total"`
	} `json:"result"`
}

// FetchFBOBatch retrieves a single page of FBO postings
func (c *OzonOrdersClient) FetchFBOBatch(ctx context.Context, since, to string, offset, limit int) ([]json.RawMessage, int, error) {
	reqBody := fboRequest{
		Dir: "ASC",
		Filter: fboFilter{
			Since: since,
			To:    to,
		},
		Limit:  limit,
		Offset: offset,
		With: fboWithOpts{
			AnalyticsData: true,
			FinancialData: true,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal fbo request: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+c.cfg.FBOEndpoint, bodyBytes)
	if err != nil {
		return nil, 0, fmt.Errorf("create fbo request: %w", err)
	}
	req.Header.Set("Client-Id", c.cfg.ClientID)
	req.Header.Set("Api-Key", c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("execute fbo request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("read fbo response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, 0, fmt.Errorf("ozon fbo api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result fboResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, 0, fmt.Errorf("unmarshal fbo response: %w", err)
	}

	return result.Result.Postings, result.Result.Total, nil
}

type fbsRequest struct {
	Dir    string      `json:"dir"`
	Filter fbsFilter   `json:"filter"`
	Limit  int         `json:"limit"`
	LastID string      `json:"last_id,omitempty"`
	With   fboWithOpts `json:"with"`
}

type fbsFilter struct {
	Since string `json:"since"`
	To    string `json:"to"`
}

type fbsResponse struct {
	Result struct {
		Postings []json.RawMessage `json:"postings"`
		HasNext  bool              `json:"has_next"`
	} `json:"result"`
}

// FetchFBSBatch retrieves a single page of FBS postings
func (c *OzonOrdersClient) FetchFBSBatch(ctx context.Context, since, to, lastID string, limit int) ([]json.RawMessage, bool, string, error) {
	reqBody := fbsRequest{
		Dir: "ASC",
		Filter: fbsFilter{
			Since: since,
			To:    to,
		},
		Limit:  limit,
		LastID: lastID,
		With: fboWithOpts{
			AnalyticsData: true,
			FinancialData: true,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, false, "", fmt.Errorf("marshal fbs request: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+c.cfg.FBSEndpoint, bodyBytes)
	if err != nil {
		return nil, false, "", fmt.Errorf("create fbs request: %w", err)
	}
	req.Header.Set("Client-Id", c.cfg.ClientID)
	req.Header.Set("Api-Key", c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, false, "", fmt.Errorf("execute fbs request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, "", fmt.Errorf("read fbs response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, false, "", fmt.Errorf("ozon fbs api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result fbsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, false, "", fmt.Errorf("unmarshal fbs response: %w", err)
	}

	postings := result.Result.Postings
	hasNext := result.Result.HasNext && len(postings) > 0
	newLastID := lastID
	if len(postings) > 0 {
		var last struct {
			PostingNumber string `json:"posting_number"`
		}
		if err := json.Unmarshal(postings[len(postings)-1], &last); err == nil {
			newLastID = last.PostingNumber
		}
	}

	return postings, hasNext, newLastID, nil
}

// FetchAllFBOOrders fetches all FBO postings within the date range, calling onBatch for each page
func (c *OzonOrdersClient) FetchAllFBOOrders(ctx context.Context, since, to string, onBatch func(orders []json.RawMessage, scheme string) error) (int, int, error) {
	limit := c.cfg.Limit
	if limit <= 0 {
		limit = 1000
	}

	offset := 0
	totalProcessed := 0
	batchesCount := 0

	for {
		postings, total, err := c.FetchFBOBatch(ctx, since, to, offset, limit)
		if err != nil {
			return totalProcessed, batchesCount, fmt.Errorf("fetch fbo batch: %w", err)
		}

		batchesCount++
		totalProcessed += len(postings)

		if len(postings) > 0 {
			if err := onBatch(postings, "FBO"); err != nil {
				return totalProcessed, batchesCount, err
			}
		}

		if len(postings) < limit || offset+len(postings) >= total {
			break
		}
		offset += len(postings)

		if c.cfg.PaginationDelayMs > 0 {
			time.Sleep(time.Duration(c.cfg.PaginationDelayMs) * time.Millisecond)
		}
	}

	return totalProcessed, batchesCount, nil
}

// FetchAllFBSOrders fetches all FBS postings within the date range, calling onBatch for each page
func (c *OzonOrdersClient) FetchAllFBSOrders(ctx context.Context, since, to string, onBatch func(orders []json.RawMessage, scheme string) error) (int, int, error) {
	limit := c.cfg.Limit
	if limit <= 0 {
		limit = 1000
	}

	lastID := ""
	totalProcessed := 0
	batchesCount := 0

	for {
		postings, hasNext, newLastID, err := c.FetchFBSBatch(ctx, since, to, lastID, limit)
		if err != nil {
			return totalProcessed, batchesCount, fmt.Errorf("fetch fbs batch: %w", err)
		}

		batchesCount++
		totalProcessed += len(postings)

		if len(postings) > 0 {
			if err := onBatch(postings, "FBS"); err != nil {
				return totalProcessed, batchesCount, err
			}
		}

		if !hasNext || len(postings) < limit || newLastID == lastID {
			break
		}
		lastID = newLastID

		if c.cfg.PaginationDelayMs > 0 {
			time.Sleep(time.Duration(c.cfg.PaginationDelayMs) * time.Millisecond)
		}
	}

	return totalProcessed, batchesCount, nil
}
