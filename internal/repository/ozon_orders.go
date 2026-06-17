// Package repository implements domain repository interfaces using PostgreSQL
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

// OzonOrderRepo is a PostgreSQL implementation of domain.OzonOrderRepository
type OzonOrderRepo struct {
	pool *pgxpool.Pool
}

// NewOzonOrderRepo returns a new OzonOrderRepo instance
func NewOzonOrderRepo(pool *pgxpool.Pool) *OzonOrderRepo {
	return &OzonOrderRepo{pool: pool}
}

// UpsertBatch inserts or updates a batch of Ozon orders
func (r *OzonOrderRepo) UpsertBatch(ctx context.Context, orders []domain.OzonOrder) (int, error) {
	if len(orders) == 0 {
		return 0, nil
	}

	const query = `
		INSERT INTO ozon_orders (
			posting_number, order_id, order_number, status,
			delivery_method_id, tpl_integration_type,
			created_at, in_process_at, shipment_date, delivering_date,
			products, analytics_data, financial_data, scheme
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		ON CONFLICT (posting_number) DO UPDATE SET
			status = EXCLUDED.status,
			products = EXCLUDED.products,
			analytics_data = EXCLUDED.analytics_data,
			financial_data = EXCLUDED.financial_data,
			updated_at = CURRENT_TIMESTAMP
	`

	batch := &pgx.Batch{}
	for _, order := range orders {
		batch.Queue(query,
			order.PostingNumber,
			order.OrderID,
			order.OrderNumber,
			order.Status,
			order.DeliveryMethodID,
			order.TplIntegrationType,
			order.CreatedAt,
			order.InProcessAt,
			order.ShipmentDate,
			order.DeliveringDate,
			order.Products,
			order.AnalyticsData,
			order.FinancialData,
			order.Scheme,
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

// GetList returns a filtered and paginated list of Ozon orders
func (r *OzonOrderRepo) GetList(ctx context.Context, filter domain.OzonOrderFilter) ([]domain.OzonOrder, int, error) {
	var (
		sb   strings.Builder
		args = make([]any, 0, 4)
		idx  = 1
	)

	sb.Grow(512)
	sb.WriteString("WHERE 1=1")

	if filter.Scheme != "" {
		fmt.Fprintf(&sb, " AND scheme = $%d", idx)
		args = append(args, filter.Scheme)
		idx++
	}
	if filter.Status != "" {
		fmt.Fprintf(&sb, " AND status = $%d", idx)
		args = append(args, filter.Status)
		idx++
	}
	if filter.From != "" {
		fmt.Fprintf(&sb, " AND created_at >= $%d", idx)
		args = append(args, filter.From)
		idx++
	}
	if filter.To != "" {
		fmt.Fprintf(&sb, " AND created_at <= $%d", idx)
		args = append(args, filter.To)
		idx++
	}

	whereClause := sb.String()

	sb.Reset()
	sb.WriteString(`SELECT posting_number, order_id, order_number, status, delivery_method_id, tpl_integration_type, created_at, in_process_at, shipment_date, delivering_date, products, analytics_data, financial_data, scheme, updated_at FROM ozon_orders `)
	sb.WriteString(whereClause)
	sb.WriteString(" ORDER BY created_at DESC")

	fmt.Fprintf(&sb, " LIMIT $%d OFFSET $%d", idx, idx+1)
	dataArgs := append(args, filter.Limit, filter.Offset)

	var orders []domain.OzonOrder
	if err := pgxscan.Select(ctx, r.pool, &orders, sb.String(), dataArgs...); err != nil {
		return nil, 0, fmt.Errorf("select ozon orders: %w", err)
	}

	sb.Reset()
	sb.WriteString("SELECT COUNT(*) FROM ozon_orders ")
	sb.WriteString(whereClause)

	var total int
	if err := pgxscan.Get(ctx, r.pool, &total, sb.String(), args...); err != nil {
		return nil, 0, fmt.Errorf("count ozon orders: %w", err)
	}

	return orders, total, nil
}

// GetStats returns aggregated statistics for Ozon orders
func (r *OzonOrderRepo) GetStats(ctx context.Context) (*domain.OzonOrderStats, error) {
	const totalQuery = `SELECT COUNT(*) FROM ozon_orders`
	const bySchemeQuery = `SELECT scheme, COUNT(*) as count FROM ozon_orders GROUP BY scheme`
	const updatedQuery = `SELECT COUNT(*) FROM ozon_orders WHERE updated_at > NOW() - INTERVAL '1 hour'`

	var total int
	if err := pgxscan.Get(ctx, r.pool, &total, totalQuery); err != nil {
		return nil, fmt.Errorf("get total ozon orders: %w", err)
	}

	rows, err := r.pool.Query(ctx, bySchemeQuery)
	if err != nil {
		return nil, fmt.Errorf("query ozon by scheme: %w", err)
	}
	defer rows.Close()

	schemeMap := make(map[string]int)
	for rows.Next() {
		var scheme string
		var count int
		if err := rows.Scan(&scheme, &count); err != nil {
			return nil, fmt.Errorf("scan scheme row: %w", err)
		}
		schemeMap[scheme] = count
	}

	var updated int
	if err := pgxscan.Get(ctx, r.pool, &updated, updatedQuery); err != nil {
		return nil, fmt.Errorf("get updated ozon orders: %w", err)
	}

	return &domain.OzonOrderStats{
		TotalOrders:     total,
		ByScheme:        schemeMap,
		UpdatedLastHour: updated,
	}, nil
}

// CountForPeriod returns the number of Ozon orders per day within a period
func (r *OzonOrderRepo) CountForPeriod(ctx context.Context, from, to string) ([]domain.DailyChartItem, error) {
	const query = `
		SELECT 
			d::date AS date,
			COALESCE(COUNT(o.posting_number), 0) AS count
		FROM generate_series($1::date, $2::date, '1 day'::interval) d
		LEFT JOIN ozon_orders o ON DATE(o.created_at) = d::date
		GROUP BY d::date
		ORDER BY d::date
	`

	var items []domain.DailyChartItem
	if err := pgxscan.Select(ctx, r.pool, &items, query, from, to); err != nil {
		return nil, fmt.Errorf("select ozon orders chart: %w", err)
	}

	return items, nil
}
