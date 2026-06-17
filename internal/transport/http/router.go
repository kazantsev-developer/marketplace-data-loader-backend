// Package http provides HTTP handlers and routing
package http

import "net/http"

// RegisterRoutes registers all API routes on the given mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/ozon/orders", h.GetOzonOrders)
	mux.HandleFunc("GET /api/ozon/orders/stats", h.GetOzonOrderStats)
	mux.HandleFunc("GET /api/charts/ozon-orders-daily", h.GetOzonOrderDailyChart)
	mux.HandleFunc("GET /api/ozon/remains", h.GetOzonRemains)
	mux.HandleFunc("GET /api/ozon/remains/stats", h.GetOzonRemainStats)

	mux.HandleFunc("GET /api/wb/orders", h.GetWbOrders)
	mux.HandleFunc("GET /api/wb/orders/stats", h.GetWbOrderStats)
	mux.HandleFunc("GET /api/charts/orders-daily", h.GetWbOrderDailyChart)
	mux.HandleFunc("GET /api/wb/remains", h.GetWbRemains)
	mux.HandleFunc("GET /api/wb/cards", h.GetWbCards)
	mux.HandleFunc("GET /api/wb/cards/stats", h.GetWbCardStats)

	mux.HandleFunc("GET /api/moysklad/stocks", h.GetMoyskladStocks)
	mux.HandleFunc("GET /api/moysklad/aggregates", h.GetMoyskladAggregates)
	mux.HandleFunc("GET /api/moysklad/stores", h.GetMoyskladStores)

	mux.HandleFunc("GET /api/sync/logs", h.GetSyncLogs)
	mux.HandleFunc("GET /api/dashboard/stats", h.GetDashboardStats)
}
