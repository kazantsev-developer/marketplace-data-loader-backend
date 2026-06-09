// Package domain предоставляет основные бизнес-сущности.
package domain

import (
	"encoding/json"
	"time"
)

// WbOrder описывает структуру заказа из API Wildberries.
type WbOrder struct {
	Srid            string    `json:"srid" db:"srid"` // Уникальный идентификатор позиции в заказе
	GNumber         string    `json:"g_number" db:"g_number"`
	Date            time.Time `json:"date" db:"date"`
	LastChangeDate  time.Time `json:"last_change_date" db:"last_change_date"`
	SupplierArticle *string   `json:"supplier_article" db:"supplier_article"`
	TechSize        *string   `json:"tech_size" db:"tech_size"`
	Barcode         *string   `json:"barcode" db:"barcode"`
	TotalPrice      float64   `json:"total_price" db:"total_price"`
	DiscountPercent int       `json:"discount_percent" db:"discount_percent"`
	WarehouseName   *string   `json:"warehouse_name" db:"warehouse_name"`
	IsCancel        bool      `json:"is_cancel" db:"is_cancel"`
	DestCityName    *string   `json:"dest_city_name" db:"dest_city_name"`
	CountryName     *string   `json:"country_name" db:"country_name"`
	OblastOkrugName *string   `json:"oblast_okrug_name" db:"oblast_okrug_name"`
	RegionName      *string   `json:"region_name" db:"region_name"`
	NmID            *int64    `json:"nm_id" db:"nm_id"`
	Category        *string   `json:"category" db:"category"`
	Brand           *string   `json:"brand" db:"brand"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// WbRemain описывает структуру остатка товара на складе Wildberries.
type WbRemain struct {
	NmID      int64     `json:"nm_id" db:"nm_id"`
	Size      string    `json:"size" db:"size"`
	Warehouse string    `json:"warehouse" db:"warehouse"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Barcode   *string   `json:"barcode" db:"barcode"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// WbCard описывает структуру карточки товара Wildberries.
type WbCard struct {
	NmID            int64           `json:"nm_id" db:"nm_id"`
	VendorCode      string          `json:"vendor_code" db:"vendor_code"`
	Brand           *string         `json:"brand" db:"brand"`
	Title           *string         `json:"title" db:"title"`
	Description     *string         `json:"description" db:"description"`
	Category        *string         `json:"category" db:"category"`
	Subject         *string         `json:"subject" db:"subject"`
	Characteristics json.RawMessage `json:"characteristics" db:"characteristics"` // Сырые JSON-данные характеристик товара
	Sizes           json.RawMessage `json:"sizes" db:"sizes"`                     // Сырые JSON-данные размерной сетки
	Photos          json.RawMessage `json:"photos" db:"photos"`                   // Сырые JSON-данные ссылок на медиафайлы
	Video           *string         `json:"video" db:"video"`
	Dimensions      json.RawMessage `json:"dimensions" db:"dimensions"` // Сырые JSON-данные габаритов товара
	Weight          *int            `json:"weight" db:"weight"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	SyncedAt        time.Time       `json:"synced_at" db:"synced_at"`
}

// OzonOrder описывает структуру отправления из API Ozon.
type OzonOrder struct {
	PostingNumber      string          `json:"posting_number" db:"posting_number"` // Уникальный номер отправления Ozon
	OrderID            *int64          `json:"order_id" db:"order_id"`
	OrderNumber        *string         `json:"order_number" db:"order_number"`
	Status             *string         `json:"status" db:"status"`
	DeliveryMethodID   *int64          `json:"delivery_method_id" db:"delivery_method_id"`
	TplIntegrationType *string         `json:"tpl_integration_type" db:"tpl_integration_type"`
	CreatedAt          *time.Time      `json:"created_at" db:"created_at"`
	InProcessAt        *time.Time      `json:"in_process_at" db:"in_process_at"`
	ShipmentDate       *time.Time      `json:"shipment_date" db:"shipment_date"`
	DeliveringDate     *time.Time      `json:"delivering_date" db:"delivering_date"`
	Products           json.RawMessage `json:"products" db:"products"`             // Сырые JSON-данные списка продуктов в заказе
	AnalyticsData      json.RawMessage `json:"analytics_data" db:"analytics_data"` // Сырые JSON-данные аналитического блока Ozon
	FinancialData      json.RawMessage `json:"financial_data" db:"financial_data"` // Сырые JSON-данные финансовых начислений и комиссий
	Scheme             *string         `json:"scheme" db:"scheme"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
}

// OzonRemain описывает структуру остатка товара на складе Ozon.
type OzonRemain struct {
	Sku              int64     `json:"sku" db:"sku"`
	ProductID        int64     `json:"product_id" db:"product_id"`
	ItemCode         *string   `json:"item_code" db:"item_code"`
	Category         *string   `json:"category" db:"category"`
	Brand            *string   `json:"brand" db:"brand"`
	Name             *string   `json:"name" db:"name"`
	FboVisibleAmount int       `json:"fbo_visible_amount" db:"fbo_visible_amount"`
	FboPresentAmount int       `json:"fbo_present_amount" db:"fbo_present_amount"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	SyncedAt         time.Time `json:"synced_at" db:"synced_at"`
}

// MsStore описывает структуру склада из ERP МойСклад.
type MsStore struct {
	UUID         string    `json:"uuid" db:"uuid"`
	Name         string    `json:"name" db:"name"`
	Code         *string   `json:"code" db:"code"`
	ExternalCode *string   `json:"external_code" db:"external_code"`
	Address      *string   `json:"address" db:"address"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	SyncedAt     time.Time `json:"synced_at" db:"synced_at"`
}

// MsSnapshot описывает зафиксированную точку исторического среза данных.
type MsSnapshot struct {
	ID          int       `json:"id" db:"id"`
	CollectedAt time.Time `json:"collected_at" db:"collected_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// MsStockDetail описывает детальные остатки товара на конкретном складе МойСклад.
type MsStockDetail struct {
	SnapshotID  int       `json:"snapshot_id" db:"snapshot_id"`
	ProductUUID string    `json:"product_uuid" db:"product_uuid"`
	StoreUUID   string    `json:"store_uuid" db:"store_uuid"`
	Stock       int       `json:"stock" db:"stock"`
	Reserve     int       `json:"reserve" db:"reserve"`
	InTransit   int       `json:"in_transit" db:"in_transit"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// MsProductTotal описывает агрегированные остатки товара по всем складам МойСклад.
type MsProductTotal struct {
	ProductUUID    string    `json:"product_uuid" db:"product_uuid"`
	Article        *string   `json:"article" db:"article"`
	Name           *string   `json:"name" db:"name"`
	TotalStock     int       `json:"total_stock" db:"total_stock"`
	TotalReserve   int       `json:"total_reserve" db:"total_reserve"`
	TotalInTransit int       `json:"total_in_transit" db:"total_in_transit"`
	SnapshotID     *int      `json:"snapshot_id" db:"snapshot_id"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// MsJobLog описывает структуру лога фонового джоба МойСклад.
type MsJobLog struct {
	ID                   int       `json:"id" db:"id"`
	StartedAt            time.Time `json:"started_at" db:"started_at"`
	Status               string    `json:"status" db:"status"`
	RecordsCount         int       `json:"records_count" db:"records_count"`
	DetailsCount         int       `json:"details_count" db:"details_count"`
	AggregatesCount      int       `json:"aggregates_count" db:"aggregates_count"`
	ErrorMessage         *string   `json:"error_message" db:"error_message"`
	ExecutionTimeSeconds *int      `json:"execution_time_seconds" db:"execution_time_seconds"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

// SyncLog описывает структуру общего лога синхронизации сущностей.
type SyncLog struct {
	ID                   int        `json:"id" db:"id"`
	SyncAt               time.Time  `json:"sync_at" db:"sync_at"`
	Status               string     `json:"status" db:"status"`
	RecordsCount         int        `json:"records_count" db:"records_count"`
	DateFrom             *time.Time `json:"date_from" db:"date_from"`
	DateTo               *time.Time `json:"date_to" db:"date_to"`
	ErrorMessage         *string    `json:"error_message" db:"error_message"`
	PagesCount           int        `json:"pages_count" db:"pages_count"`
	ExecutionTimeSeconds int        `json:"execution_time_seconds" db:"execution_time_seconds"`
	EntityType           string     `json:"entity_type" db:"entity_type"`
}

// OzonSyncLog описывает структуру специализированного лога синхронизации Ozon.
type OzonSyncLog struct {
	ID           int        `json:"id" db:"id"`
	SyncAt       time.Time  `json:"sync_at" db:"sync_at"`
	Status       string     `json:"status" db:"status"`
	Scheme       string     `json:"scheme" db:"scheme"`
	RecordsCount int        `json:"records_count" db:"records_count"`
	DateFrom     *time.Time `json:"date_from" db:"date_from"`
}

type DailyChartItem struct {
	Date  string
	Count int
}

type SyncCursorState struct {
	UpdatedAt string
	NmID      int64
}
