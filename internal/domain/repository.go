// Package domain defines core business interfaces and filters
package domain

import "context"

// OzonOrderRepository provides an interface for managing Ozon orders
type OzonOrderRepository interface {
	UpsertBatch(ctx context.Context, orders []OzonOrder) (int, error)
	GetList(ctx context.Context, filter OzonOrderFilter) ([]OzonOrder, int, error)
	GetStats(ctx context.Context) (*OzonOrderStats, error)
	CountForPeriod(ctx context.Context, from, to string) ([]DailyChartItem, error)
}

// OzonOrderFilter contains parameters for filtering Ozon orders
type OzonOrderFilter struct {
	Scheme string
	Status string
	From   string
	To     string
	Limit  int
	Offset int
}

// OzonOrderStats holds aggregated statistics for Ozon orders
type OzonOrderStats struct {
	TotalOrders     int            `json:"total_orders"`
	ByScheme        map[string]int `json:"by_scheme"`
	UpdatedLastHour int            `json:"updated_last_hour"`
}

// OzonRemainRepository provides an interface for managing Ozon stocks
type OzonRemainRepository interface {
	UpsertBatch(ctx context.Context, remains []OzonRemain) (int, error)
	GetAll(ctx context.Context, brand, search string) ([]OzonRemain, error)
	ResetStale(ctx context.Context, before string) (int, error)
	GetStats(ctx context.Context) (*OzonRemainStats, error)
}

// OzonRemainStats holds statistics for Ozon stocks
type OzonRemainStats struct {
	TotalProducts   int         `json:"total_products"`
	TotalVisible    int         `json:"total_visible"`
	TotalPresent    int         `json:"total_present"`
	TopBrands       []BrandStat `json:"top_brands"`
	UpdatedLastHour int         `json:"updated_last_hour"`
}

// BrandStat holds product statistics aggregated by brand
type BrandStat struct {
	Brand    string `json:"brand"`
	Products int    `json:"products"`
	Visible  int    `json:"visible"`
}

// WbOrderRepository provides an interface for managing Wildberries orders
type WbOrderRepository interface {
	UpsertBatch(ctx context.Context, orders []WbOrder) (int, error)
	GetList(ctx context.Context, filter OrderFilter) ([]WbOrder, int, error)
	GetStats(ctx context.Context) (*WbOrderStats, error)
	CountForPeriod(ctx context.Context, from, to string) ([]DailyChartItem, error)
}

// OrderFilter contains parameters for filtering Wildberries order lists
type OrderFilter struct {
	From   string
	To     string
	Limit  int
	Offset int
}

// WbOrderStats holds aggregated statistics for Wildberries orders
type WbOrderStats struct {
	TotalOrders     int     `json:"total_orders"`
	CancelledOrders int     `json:"cancelled_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	UniqueProducts  int     `json:"unique_products"`
}

// WbRemainRepository provides an interface for managing Wildberries stocks
type WbRemainRepository interface {
	UpsertBatch(ctx context.Context, remains []WbRemain) (int, error)
	GetAll(ctx context.Context, warehouse, search string) ([]WbRemain, error)
}

// WbCardRepository provides an interface for managing Wildberries product cards
type WbCardRepository interface {
	UpsertBatch(ctx context.Context, cards []WbCard) (int, error)
	GetList(ctx context.Context, search string, limit, offset int) ([]WbCard, int, error)
	GetStats(ctx context.Context) (*WbCardStats, error)
	GetCursor(ctx context.Context) (*SyncCursorState, error)
	SaveCursor(ctx context.Context, updatedAt string, nmID int64) error
}

// WbCardStats holds statistics for Wildberries product cards
type WbCardStats struct {
	TotalCards      int `json:"total_cards"`
	UpdatedLastHour int `json:"updated_last_hour"`
}

// MoyskladRepository provides an interface for managing MoySklad ERP data
type MoyskladRepository interface {
	UpsertStores(ctx context.Context, stores []MsStore) (int, error)
	GetStores(ctx context.Context) ([]MsStore, error)
	CreateSnapshot(ctx context.Context) (int, error)
	InsertStockDetails(ctx context.Context, details []MsStockDetail) (int, error)
	UpsertProductTotals(ctx context.Context, totals []MsProductTotal) (int, error)
	GetStockDetails(ctx context.Context, productUUID, storeUUID string) ([]MsStockDetail, error)
	GetProductTotals(ctx context.Context) ([]MsProductTotal, error)
	GetStockTotal(ctx context.Context) (int, error)
}

// SyncLogRepository provides an interface for managing synchronization logs
type SyncLogRepository interface {
	Insert(ctx context.Context, log SyncLog) (int, error)
	GetList(ctx context.Context, entityType string, limit int) ([]SyncLog, error)
	InsertOzonLog(ctx context.Context, log OzonSyncLog) (int, error)
	InsertMsJobLog(ctx context.Context, log MsJobLog) (int, error)
}

// SyncStats holds aggregated statistics for synchronization sessions
type SyncStats struct {
	Last24h     int `json:"last_24h"`
	SuccessRate int `json:"success_rate"`
}
