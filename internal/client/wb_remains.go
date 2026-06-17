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

// WbRemainsClient is an HTTP client for the Wildberries Warehouse Remains API
type WbRemainsClient struct {
	cfg    config.WbConfig
	client *retryablehttp.Client
}

// NewWbRemainsClient returns a new WbRemainsClient instance with configured retry policy
func NewWbRemainsClient(cfg config.WbConfig) *WbRemainsClient {
	rc := retryablehttp.NewClient()
	rc.Logger = nil
	rc.RetryMax = cfg.MaxRetries
	rc.RetryWaitMin = time.Duration(cfg.PaginationDelayMs) * time.Millisecond
	rc.RetryWaitMax = time.Duration(cfg.PaginationDelayMs) * time.Millisecond
	rc.CheckRetry = retryPolicy
	rc.HTTPClient.Timeout = cfg.Timeout

	return &WbRemainsClient{
		cfg:    cfg,
		client: rc,
	}
}

// CreateRemainsReport requests generation of a new warehouse remains report and returns the task ID for polling
func (c *WbRemainsClient) CreateRemainsReport(ctx context.Context) (string, error) {
	reqURL, err := url.Parse("https://seller-analytics-api.wildberries.ru/api/v1/warehouse_remains")
	if err != nil {
		return "", fmt.Errorf("parse remains report url: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create remains request: %w", err)
	}
	req.Header.Set("Authorization", c.cfg.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute remains request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read remains response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("wb remains api returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			TaskID string `json:"taskId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("unmarshal remains task id: %w", err)
	}

	if result.Data.TaskID == "" {
		return "", fmt.Errorf("empty taskId in response: %s", string(body))
	}

	return result.Data.TaskID, nil
}

// CheckReportStatus returns the current status of the report generation task.
// Possible statuses: "processing", "done", "error"
func (c *WbRemainsClient) CheckReportStatus(ctx context.Context, taskID string) (string, error) {
	reqURL, err := url.Parse(fmt.Sprintf("https://seller-analytics-api.wildberries.ru/api/v1/warehouse_remains/tasks/%s/status", taskID))
	if err != nil {
		return "", fmt.Errorf("parse status url: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create status request: %w", err)
	}
	req.Header.Set("Authorization", c.cfg.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("check status request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read status response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("wb status api returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("unmarshal report status: %w", err)
	}

	status := result.Status
	if status == "" {
		status = result.Data.Status
	}
	if status == "" {
		return "", fmt.Errorf("unknown status format: %s", string(body))
	}

	return status, nil
}

// DownloadRemainsReport retrieves the generated report data as a JSON byte slice
func (c *WbRemainsClient) DownloadRemainsReport(ctx context.Context, taskID string) ([]byte, error) {
	reqURL, err := url.Parse(fmt.Sprintf("https://seller-analytics-api.wildberries.ru/api/v1/warehouse_remains/tasks/%s/download", taskID))
	if err != nil {
		return nil, fmt.Errorf("parse download url: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create download request: %w", err)
	}
	req.Header.Set("Authorization", c.cfg.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download report request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read download response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("wb download api returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
