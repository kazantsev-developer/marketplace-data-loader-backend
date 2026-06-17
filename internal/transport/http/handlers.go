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
	OrderRepo      domain.WbOrderRepository
	RemainRepo     domain.WbRemainRepository
	CardRepo       domain.WbCardRepository
	LogRepo        domain.SyncLogRepository
	MoySkladRepo   domain.MoyskladRepository
	OzonOrderRepo  domain.OzonOrderRepository
	OzonRemainRepo domain.OzonRemainRepository
}

// NewHandler returns a new Handler instance
func NewHandler(
	orderRepo domain.WbOrderRepository,
	remainRepo domain.WbRemainRepository,
	cardRepo domain.WbCardRepository,
	logRepo domain.SyncLogRepository,
	msRepo domain.MoyskladRepository,
	ozonOrderRepo domain.OzonOrderRepository,
	ozonRemainRepo domain.OzonRemainRepository,
) *Handler {
	return &Handler{
		OrderRepo:      orderRepo,
		RemainRepo:     remainRepo,
		CardRepo:       cardRepo,
		LogRepo:        logRepo,
		MoySkladRepo:   msRepo,
		OzonOrderRepo:  ozonOrderRepo,
		OzonRemainRepo: ozonRemainRepo,
	}
}

// GetWbOrders handles GET /api/wb/orders
func (h *Handler) GetWbOrders(w http.ResponseWriter, r *http.Request) {
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

// GetWbOrderStats handles GET /api/wb/orders/stats
func (h *Handler) GetWbOrderStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.OrderRepo.GetStats(r.Context())
	if err != nil {
		writeError(w, "get order stats", err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// GetWbOrderDailyChart handles GET /api/charts/orders-daily
func (h *Handler) GetWbOrderDailyChart(w http.ResponseWriter, r *http.Request) {
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

// GetWbRemains handles GET /api/wb/remains
func (h *Handler) GetWbRemains(w http.ResponseWriter, r *http.Request) {
	warehouse := r.URL.Query().Get("warehouse")
	search := r.URL.Query().Get("search")

	remains, err := h.RemainRepo.GetAll(r.Context(), warehouse, search)
	if err != nil {
		writeError(w, "fetch remains", err)
		return
	}
	writeJSON(w, http.StatusOK, remains)
}

// GetWbCards handles GET /api/wb/cards
func (h *Handler) GetWbCards(w http.ResponseWriter, r *http.Request) {
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

// GetWbCardStats handles GET /api/wb/cards/stats
func (h *Handler) GetWbCardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.CardRepo.GetStats(r.Context())
	if err != nil {
		writeError(w, "get card stats", err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// GetMoyskladStocks returns stock details filtered by product and store UUID
func (h *Handler) GetMoyskladStocks(w http.ResponseWriter, r *http.Request) {
	productUUID := r.URL.Query().Get("product_uuid")
	storeUUID := r.URL.Query().Get("store_uuid")

	details, err := h.MoySkladRepo.GetStockDetails(r.Context(), productUUID, storeUUID)
	if err != nil {
		writeError(w, "fetch moysklad stocks", err)
		return
	}
	writeJSON(w, http.StatusOK, details)
}

// GetMoyskladAggregates returns product totals ordered by stock descending
func (h *Handler) GetMoyskladAggregates(w http.ResponseWriter, r *http.Request) {
	totals, err := h.MoySkladRepo.GetProductTotals(r.Context())
	if err != nil {
		writeError(w, "fetch moysklad aggregates", err)
		return
	}
	writeJSON(w, http.StatusOK, totals)
}

// GetMoyskladStores returns all MoySklad warehouses
func (h *Handler) GetMoyskladStores(w http.ResponseWriter, r *http.Request) {
	stores, err := h.MoySkladRepo.GetStores(r.Context())
	if err != nil {
		writeError(w, "fetch moysklad stores", err)
		return
	}
	writeJSON(w, http.StatusOK, stores)
}

// GetOzonOrders handles GET /api/ozon/orders
func (h *Handler) GetOzonOrders(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r, 100, 0)
	filter := domain.OzonOrderFilter{
		Scheme: r.URL.Query().Get("scheme"),
		Status: r.URL.Query().Get("status"),
		From:   r.URL.Query().Get("from"),
		To:     r.URL.Query().Get("to"),
		Limit:  limit,
		Offset: offset,
	}

	orders, total, err := h.OzonOrderRepo.GetList(r.Context(), filter)
	if err != nil {
		writeError(w, "fetch ozon orders", err)
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

// GetOzonOrderStats handles GET /api/ozon/orders/stats
func (h *Handler) GetOzonOrderStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.OzonOrderRepo.GetStats(r.Context())
	if err != nil {
		writeError(w, "get ozon order stats", err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// GetOzonOrderDailyChart handles GET /api/charts/ozon-orders-daily
func (h *Handler) GetOzonOrderDailyChart(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		writeError(w, "validate date range", domain.ErrBadRequest("from and to query parameters are required"))
		return
	}

	items, err := h.OzonOrderRepo.CountForPeriod(r.Context(), from, to)
	if err != nil {
		writeError(w, "count ozon orders for period", err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// GetOzonRemains handles GET /api/ozon/remains
func (h *Handler) GetOzonRemains(w http.ResponseWriter, r *http.Request) {
	brand := r.URL.Query().Get("brand")
	search := r.URL.Query().Get("search")

	remains, err := h.OzonRemainRepo.GetAll(r.Context(), brand, search)
	if err != nil {
		writeError(w, "fetch ozon remains", err)
		return
	}
	writeJSON(w, http.StatusOK, remains)
}

// GetOzonRemainStats handles GET /api/ozon/remains/stats
func (h *Handler) GetOzonRemainStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.OzonRemainRepo.GetStats(r.Context())
	if err != nil {
		writeError(w, "get ozon remain stats", err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// GetSyncLogs handles GET /api/sync/logs
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

// GetDashboardStats returns aggregated statistics for the dashboard
func (h *Handler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]any{
		"status": "ok",
	}
	writeJSON(w, http.StatusOK, stats)
}

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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, context string, err error) {
	status := http.StatusInternalServerError
	if domain.IsBadRequest(err) {
		status = http.StatusBadRequest
	}
	http.Error(w, context+": "+err.Error(), status)
}
