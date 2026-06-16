// Package http provides HTTP handlers for the REST API
package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/domain"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	OrderRepo  domain.WbOrderRepository
	RemainRepo domain.WbRemainRepository
	CardRepo   domain.WbCardRepository
	LogRepo    domain.SyncLogRepository
}

// NewHandler returns a new Handler instance
func NewHandler(
	orderRepo domain.WbOrderRepository,
	remainRepo domain.WbRemainRepository,
	cardRepo domain.WbCardRepository,
	logRepo domain.SyncLogRepository,
) *Handler {
	return &Handler{
		OrderRepo:  orderRepo,
		RemainRepo: remainRepo,
		CardRepo:   cardRepo,
		LogRepo:    logRepo,
	}
}

// GetOrders handles GET /api/wb/orders
func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit, offset := parsePagination(r, 100, 0)
	filter := domain.OrderFilter{
		From:   r.URL.Query().Get("from"),
		To:     r.URL.Query().Get("to"),
		Limit:  limit,
		Offset: offset,
	}

	orders, total, err := h.OrderRepo.GetList(ctx, filter)
	if err != nil {
		writeError(w, "fetch orders", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": orders,
		"pagination": map[string]int{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetOrderStats handles GET /api/wb/orders/stats
func (h *Handler) GetOrderStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.OrderRepo.GetStats(r.Context())
	if err != nil {
		writeError(w, "get order stats", err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// GetOrderDailyChart handles GET /api/charts/orders-daily?from=...&to=...
func (h *Handler) GetOrderDailyChart(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		writeError(w, "validate date range", domain.ErrBadRequest("from and to query parameters are required"))
		return
	}

	items, err := h.OrderRepo.CountForPeriod(r.Context(), from, to)
	if err != nil {
		writeError(w, "count orders for period", err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// GetRemains handles GET /api/wb/remains
func (h *Handler) GetRemains(w http.ResponseWriter, r *http.Request) {
	warehouse := r.URL.Query().Get("warehouse")
	search := r.URL.Query().Get("search")

	remains, err := h.RemainRepo.GetAll(r.Context(), warehouse, search)
	if err != nil {
		writeError(w, "fetch remains", err)
		return
	}
	writeJSON(w, http.StatusOK, remains)
}

// GetCards handles GET /api/wb/cards
func (h *Handler) GetCards(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r, 50, 0)
	search := r.URL.Query().Get("search")

	cards, total, err := h.CardRepo.GetList(r.Context(), search, limit, offset)
	if err != nil {
		writeError(w, "fetch cards", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": cards,
		"pagination": map[string]int{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetCardStats handles GET /api/wb/cards/stats
func (h *Handler) GetCardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.CardRepo.GetStats(r.Context())
	if err != nil {
		writeError(w, "get card stats", err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// GetSyncLogs handles GET /api/sync/logs?entity=...&limit=...
func (h *Handler) GetSyncLogs(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	logs, err := h.LogRepo.GetList(r.Context(), entityType, limit)
	if err != nil {
		writeError(w, "fetch sync logs", err)
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

// parsePagination extracts limit and offset from query string
func parsePagination(r *http.Request, defaultLimit, defaultOffset int) (int, int) {
	limit := defaultLimit
	offset := defaultOffset

	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && o >= 0 {
		offset = o
	}
	return limit, offset
}

// writeJSON serializes v as JSON and writes it to the response
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// writeError logs and sends a JSON error response
func writeError(w http.ResponseWriter, context string, err error) {
	status := http.StatusInternalServerError
	if domain.IsBadRequest(err) {
		status = http.StatusBadRequest
	}
	http.Error(w, context+": "+err.Error(), status)
}
