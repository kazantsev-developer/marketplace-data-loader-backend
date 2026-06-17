// Package client implements HTTP clients for marketplace APIs
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/config"
)

// MoyskladClient is an HTTP client for the MoySklad JSON API
type MoyskladClient struct {
	cfg        config.MoyskladConfig
	client     *retryablehttp.Client
	mu         sync.Mutex
	timestamps []time.Time
}

// NewMoyskladClient returns a new MoyskladClient instance with configured retry policy
func NewMoyskladClient(cfg config.MoyskladConfig) *MoyskladClient {
	rc := retryablehttp.NewClient()
	rc.Logger = nil
	rc.RetryMax = cfg.MaxRetries
	rc.RetryWaitMin = time.Duration(cfg.RetryDelayMs) * time.Millisecond
	rc.RetryWaitMax = time.Duration(cfg.RetryDelayMs*2) * time.Millisecond
	rc.CheckRetry = retryPolicy
	rc.HTTPClient.Timeout = cfg.Timeout

	return &MoyskladClient{
		cfg:    cfg,
		client: rc,
	}
}

func (c *MoyskladClient) checkRateLimit(ctx context.Context, isHeavy bool) error {
	var waitDuration time.Duration

	c.mu.Lock()
	now := time.Now()
	limit := 45
	if isHeavy {
		limit = 5
	}

	valid := c.timestamps[:0]
	for _, t := range c.timestamps {
		if now.Sub(t) < time.Minute {
			valid = append(valid, t)
		}
	}
	c.timestamps = valid

	if len(c.timestamps) >= limit {
		waitDuration = time.Minute - now.Sub(c.timestamps[0])
	}
	c.timestamps = append(c.timestamps, now)
	c.mu.Unlock()

	if waitDuration > 0 {
		select {
		case <-time.After(waitDuration):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (c *MoyskladClient) doRequest(ctx context.Context, method, path string, params url.Values, isHeavy bool) ([]byte, error) {
	if err := c.checkRateLimit(ctx, isHeavy); err != nil {
		return nil, err
	}

	reqURL, err := url.Parse(c.cfg.BaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	reqURL.RawQuery = params.Encode()

	req, err := retryablehttp.NewRequestWithContext(ctx, method, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	req.Header.Set("Accept", "application/json;charset=utf-8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("moysklad api returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// FetchStores retrieves all warehouses from MoySklad
func (c *MoyskladClient) FetchStores(ctx context.Context) ([]json.RawMessage, error) {
	allStores := make([]json.RawMessage, 0)
	offset := 0
	limit := 1000

	for {
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", limit))
		params.Set("offset", fmt.Sprintf("%d", offset))

		body, err := c.doRequest(ctx, http.MethodGet, "/entity/store", params, false)
		if err != nil {
			return nil, fmt.Errorf("fetch stores: %w", err)
		}

		var result struct {
			Rows []json.RawMessage `json:"rows"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("unmarshal stores: %w", err)
		}

		allStores = append(allStores, result.Rows...)
		if len(result.Rows) < limit {
			break
		}
		offset += limit

		select {
		case <-time.After(time.Duration(c.cfg.PaginationDelayMs) * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return allStores, nil
}

// FetchStockByStorePage retrieves a single page of the stock-by-store report
func (c *MoyskladClient) FetchStockByStorePage(ctx context.Context, offset, limit int) ([]json.RawMessage, int64, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))
	params.Set("stockMode", "byStore")

	body, err := c.doRequest(ctx, http.MethodGet, "/report/stock/bystore", params, true)
	if err != nil {
		return nil, 0, err
	}

	var result struct {
		Rows []json.RawMessage `json:"rows"`
		Meta struct {
			Size int64 `json:"size"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 0, fmt.Errorf("unmarshal stock report: %w", err)
	}

	return result.Rows, result.Meta.Size, nil
}

// FetchAllStockByStore retrieves the full stock report
func (c *MoyskladClient) FetchAllStockByStore(ctx context.Context) ([]json.RawMessage, error) {
	allRows := make([]json.RawMessage, 0)
	limit := 1000
	offset := 0

	for {
		rows, _, err := c.FetchStockByStorePage(ctx, offset, limit)
		if err != nil {
			return nil, fmt.Errorf("fetch stock page: %w", err)
		}

		allRows = append(allRows, rows...)
		if len(rows) < limit {
			break
		}
		offset += limit

		select {
		case <-time.After(time.Duration(c.cfg.HeavyRequestDelayMs) * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return allRows, nil
}
