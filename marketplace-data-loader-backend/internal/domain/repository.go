// Package domain определяет интерфейсы репозиториев.
package domain

import "context"

// WbOrderRepository предоставляет интерфейс для управления заказами Wildberries.
type WbOrderRepository interface {
	// UpsertBatch выполняет пакетное сохранение заказов.
	UpsertBatch(ctx context.Context, orders []WbOrder) (int, error)
	GetList(ctx context.Context, filter OrderFilter) ([]WbOrder, int, error)
	GetStats(ctx context.Context) (*WbOrderStats, error)
	CountForPeriod(ctx context.Context, from, to string) ([]DailyChartItem, error)
}

// OrderFilter содержит параметры фильтрации списков заказов.
type OrderFilter struct {
	From   string
	To     string
	Limit  int
	Offset int
}

// WbOrderStats содержит агрегированную статистику по заказам Wildberries.
type WbOrderStats struct {
	TotalOrders     int     `json:"total_orders"`
	CancelledOrders int     `json:"cancelled_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	UniqueProducts  int     `json:"unique_products"`
}

// WbRemainRepository предоставляет интерфейс для управления остатками Wildberries.
type WbRemainRepository interface {
	UpsertBatch(ctx context.Context, remains []WbRemain) (int, error)
	GetAll(ctx context.Context, warehouse, search string) ([]WbRemain, error)
}

// WbCardRepository предоставляет интерфейс для работы с карточками товаров Wildberries.
type WbCardRepository interface {
	UpsertBatch(ctx context.Context, cards []WbCard) (int, error)
	GetList(ctx context.Context, search string, limit, offset int) ([]WbCard, int, error)
	GetStats(ctx context.Context) (*WbCardStats, error)
	// GetCursor возвращает состояние указателя для инкрементальной синхронизации.
	GetCursor(ctx context.Context) (*SyncCursorState, error)
	SaveCursor(ctx context.Context, updatedAt string, nmID int64) error
}

// WbCardStats содержит статистику по карточкам товаров Wildberries.
type WbCardStats struct {
	TotalCards      int `json:"total_cards"`
	UpdatedLastHour int `json:"updated_last_hour"`
}

// OzonOrderRepository предоставляет интерфейс для управления заказами Ozon.
type OzonOrderRepository interface {
	UpsertBatch(ctx context.Context, orders []OzonOrder) (int, error)
	GetList(ctx context.Context, filter OzonOrderFilter) ([]OzonOrder, int, error)
	GetStats(ctx context.Context) (*OzonOrderStats, error)
	CountForPeriod(ctx context.Context, from, to string) ([]DailyChartItem, error)
}

// OzonOrderFilter содержит параметры фильтрации заказов Ozon.
type OzonOrderFilter struct {
	Scheme string
	Status string
	From   string
	To     string
	Limit  int
	Offset int
}

// OzonOrderStats содержит агрегированную статистику по заказам Ozon.
type OzonOrderStats struct {
	TotalOrders     int            `json:"total_orders"`
	ByScheme        map[string]int `json:"by_scheme"`
	UpdatedLastHour int            `json:"updated_last_hour"`
}

// OzonRemainRepository предоставляет интерфейс для управления остатками Ozon.
type OzonRemainRepository interface {
	UpsertBatch(ctx context.Context, remains []OzonRemain) (int, error)
	GetAll(ctx context.Context, brand, search string) ([]OzonRemain, error)
	// ResetStale обнуляет остатки, которые не были обновлены в текущей сессии синхронизации.
	ResetStale(ctx context.Context, before string) (int, error)
	GetStats(ctx context.Context) (*OzonRemainStats, error)
}

// OzonRemainStats содержит статистику по остаткам Ozon.
type OzonRemainStats struct {
	TotalProducts   int         `json:"total_products"`
	TotalVisible    int         `json:"total_visible"`
	TotalPresent    int         `json:"total_present"`
	TopBrands       []BrandStat `json:"top_brands"`
	UpdatedLastHour int         `json:"updated_last_hour"`
}

// BrandStat содержит статистику продуктов в разрезе бренда.
type BrandStat struct {
	Brand    string `json:"brand"`
	Products int    `json:"products"`
	Visible  int    `json:"visible"`
}

// MoyskladRepository предоставляет интерфейс для управления данными МойСклад.
type MoyskladRepository interface {
	UpsertStores(ctx context.Context, stores []MsStore) (int, error)
	GetStores(ctx context.Context) ([]MsStore, error)
	// CreateSnapshot создает исторический срез остатков.
	CreateSnapshot(ctx context.Context) (int, error)
	InsertStockDetails(ctx context.Context, details []MsStockDetail) (int, error)
	UpsertProductTotals(ctx context.Context, totals []MsProductTotal) (int, error)
	GetStockDetails(ctx context.Context, productUUID, storeUUID string) ([]MsStockDetail, error)
	GetProductTotals(ctx context.Context) ([]MsProductTotal, error)
	GetStockTotal(ctx context.Context) (int, error)
}

// SyncLogRepository предоставляет интерфейс для управления логами синхронизации.
type SyncLogRepository interface {
	Insert(ctx context.Context, log SyncLog) (int, error)
	GetList(ctx context.Context, entityType string, limit int) ([]SyncLog, error)
	InsertOzonLog(ctx context.Context, log OzonSyncLog) (int, error)
	InsertMsJobLog(ctx context.Context, log MsJobLog) (int, error)
}

// SyncStats содержит агрегированную статистику по сессиям синхронизации.
type SyncStats struct {
	Last24h     int `json:"last_24h"`
	SuccessRate int `json:"success_rate"`
}
