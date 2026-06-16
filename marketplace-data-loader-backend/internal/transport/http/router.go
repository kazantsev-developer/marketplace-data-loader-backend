// Package http provides HTTP handlers and routing
package http

import "net/http"

// RegisterRoutes registers all Wildberries API routes on the given mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/wb/orders", h.GetOrders)
	mux.HandleFunc("GET /api/wb/orders/stats", h.GetOrderStats)
	mux.HandleFunc("GET /api/charts/orders-daily", h.GetOrderDailyChart)
	mux.HandleFunc("GET /api/wb/remains", h.GetRemains)
	mux.HandleFunc("GET /api/wb/cards", h.GetCards)
	mux.HandleFunc("GET /api/wb/cards/stats", h.GetCardStats)
	mux.HandleFunc("GET /api/sync/logs", h.GetSyncLogs)
}
