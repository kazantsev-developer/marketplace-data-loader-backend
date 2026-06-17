// Package client implements HTTP clients for marketplace APIs
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/config"
)

// WbClient is an HTTP client for the Wildberries Statistics API
type WbClient struct {
	cfg    config.WbConfig
	client *retryablehttp.Client
}

// NewWbClient returns a new WbClient instance with configured retry policy
func NewWbClient(cfg config.WbConfig) *WbClient {
	rc := retryablehttp.NewClient()
	rc.Logger = nil
	rc.RetryMax = cfg.MaxRetries
	rc.RetryWaitMin = time.Duration(cfg.PaginationDelayMs) * time.Millisecond
	rc.RetryWaitMax = time.Duration(cfg.PaginationDelayMs) * time.Millisecond
	rc.CheckRetry = retryPolicy
	rc.HTTPClient.Timeout = cfg.Timeout

	return &WbClient{
		cfg:    cfg,
		client: rc,
	}
}

// FetchOrders retrieves a single page of orders from the WB API
func (c *WbClient) FetchOrders(ctx context.Context, dateFrom string) ([]json.RawMessage, error) {
	reqURL, err := url.Parse(c.cfg.BaseURL + c.cfg.OrdersEndpoint)
	if err != nil {
		return nil, fmt.Errorf("parse wb orders url: %w", err)
	}

	q := reqURL.Query()
	q.Set("dateFrom", dateFrom)
	q.Set("flag", fmt.Sprintf("%d", c.cfg.Flag))
	reqURL.RawQuery = q.Encode()

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create wb request: %w", err)
	}
	req.Header.Set("Authorization", c.cfg.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute wb request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read wb response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("wb api returned status %d: %s", resp.StatusCode, string(body))
	}

	var orders []json.RawMessage
	if err := json.Unmarshal(body, &orders); err != nil {
		return nil, fmt.Errorf("unmarshal wb orders: %w", err)
	}

	return orders, nil
}

// retryPolicy defines when the client should retry a request
func retryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if err != nil {
		return true, nil
	}
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return true, nil
	}
	return false, nil
}
