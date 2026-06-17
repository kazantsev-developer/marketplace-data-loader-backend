// Package repository provides implementations of domain repository interfaces using PostgreSQL
package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/domain"
)

// WbOrderRepo is a PostgreSQL implementation of domain.WbOrderRepository
type WbOrderRepo struct {
	pool *pgxpool.Pool
}

// NewWbOrderRepo returns a new WbOrderRepo instance
func NewWbOrderRepo(pool *pgxpool.Pool) *WbOrderRepo {
	return &WbOrderRepo{pool: pool}
}

// UpsertBatch inserts or updates a batch of WB orders
func (r *WbOrderRepo) UpsertBatch(ctx context.Context, orders []domain.WbOrder) (int, error) {
	if len(orders) == 0 {
		return 0, nil
	}

	const query = `
		INSERT INTO wb_orders (
			srid, g_number, date, last_change_date, supplier_article,
			tech_size, barcode, total_price, discount_percent, warehouse_name,
			is_cancel, dest_city_name, country_name, oblast_okrug_name, region_name,
			nm_id, category, brand
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18
		)
		ON CONFLICT (srid) DO UPDATE SET
			last_change_date = EXCLUDED.last_change_date,
			is_cancel = EXCLUDED.is_cancel,
			total_price = EXCLUDED.total_price,
			updated_at = CURRENT_TIMESTAMP
	`

	batch := &pgx.Batch{}
	for _, order := range orders {
		batch.Queue(query,
			order.Srid,
			order.GNumber,
			order.Date,
			order.LastChangeDate,
			order.SupplierArticle,
			order.TechSize,
			order.Barcode,
			order.TotalPrice,
			order.DiscountPercent,
			order.WarehouseName,
			order.IsCancel,
			order.DestCityName,
			order.CountryName,
			order.OblastOkrugName,
			order.RegionName,
			order.NmID,
			order.Category,
			order.Brand,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	var totalRows int64
	for i := range orders {
		ct, err := br.Exec()
		if err != nil {
			return 0, fmt.Errorf("execute batch item %d: %w", i, err)
		}
		totalRows += ct.RowsAffected()
	}

	return int(totalRows), nil
}

// GetList returns a filtered and paginated list of WB orders
func (r *WbOrderRepo) GetList(ctx context.Context, filter domain.OrderFilter) ([]domain.WbOrder, int, error) {
	var (
		sb   strings.Builder
		args = make([]any, 0, 4)
		idx  = 1
	)

	sb.Grow(512)
	sb.WriteString("WHERE 1=1")

	if filter.From != "" {
		fmt.Fprintf(&sb, " AND date >= $%d", idx)
		args = append(args, filter.From)
		idx++
	}
	if filter.To != "" {
		fmt.Fprintf(&sb, " AND date <= $%d", idx)
		args = append(args, filter.To)
		idx++
	}

	whereClause := sb.String()

	sb.Reset()
	sb.WriteString(`
		SELECT srid, g_number, date, last_change_date, supplier_article,
		       tech_size, barcode, total_price, discount_percent, warehouse_name,
		       is_cancel, dest_city_name, country_name, oblast_okrug_name, region_name,
		       nm_id, category, brand, created_at, updated_at
		FROM wb_orders `)
	sb.WriteString(whereClause)
	sb.WriteString(" ORDER BY date DESC")

	fmt.Fprintf(&sb, " LIMIT $%d OFFSET $%d", idx, idx+1)
	dataArgs := append(args, filter.Limit, filter.Offset)

	var orders []domain.WbOrder
	if err := pgxscan.Select(ctx, r.pool, &orders, sb.String(), dataArgs...); err != nil {
		return nil, 0, fmt.Errorf("select orders: %w", err)
	}

	sb.Reset()
	sb.WriteString("SELECT COUNT(*) FROM wb_orders ")
	sb.WriteString(whereClause)

	var total int
	if err := pgxscan.Get(ctx, r.pool, &total, sb.String(), args...); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	return orders, total, nil
}

// GetStats returns aggregated statistics for WB orders
func (r *WbOrderRepo) GetStats(ctx context.Context) (*domain.WbOrderStats, error) {
	const query = `
		SELECT 
			COUNT(*) as total_orders,
			SUM(CASE WHEN is_cancel THEN 1 ELSE 0 END) as cancelled_orders,
			SUM(total_price) as total_revenue,
			COUNT(DISTINCT nm_id) as unique_products
		FROM wb_orders
	`

	var stats domain.WbOrderStats
	if err := pgxscan.Get(ctx, r.pool, &stats, query); err != nil {
		return nil, fmt.Errorf("get order stats: %w", err)
	}

	return &stats, nil
}

// CountForPeriod returns the number of orders per day within a period
func (r *WbOrderRepo) CountForPeriod(ctx context.Context, from, to string) ([]domain.DailyChartItem, error) {
	const query = `
		SELECT 
			d::date AS date,
			COALESCE(COUNT(wb.srid), 0) AS count
		FROM generate_series($1::date, $2::date, '1 day'::interval) d
		LEFT JOIN wb_orders wb ON DATE(wb.date) = d::date
		GROUP BY d::date
		ORDER BY d::date
	`

	var items []domain.DailyChartItem
	if err := pgxscan.Select(ctx, r.pool, &items, query, from, to); err != nil {
		return nil, fmt.Errorf("select orders chart: %w", err)
	}

	return items, nil
}
