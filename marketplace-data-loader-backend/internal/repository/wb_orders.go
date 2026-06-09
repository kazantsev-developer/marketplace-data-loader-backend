// Package repository provides implementations of domain repository interfaces using PostgreSQL.
package repository

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/domain"
)

// WbOrderRepo реализует domain.WbOrderRepository с использованием pgxpool.
type WbOrderRepo struct {
	pool *pgxpool.Pool
}

// NewWbOrderRepo создаёт новый экземпляр WbOrderRepo.
func NewWbOrderRepo(pool *pgxpool.Pool) *WbOrderRepo {
	return &WbOrderRepo{pool: pool}
}

// UpsertBatch выполняет пакетную вставку или обновление заказов.
func (r *WbOrderRepo) UpsertBatch(ctx context.Context, orders []domain.WbOrder) (int, error) {
	if len(orders) == 0 {
		return 0, nil
	}

	// Строим пакетный запрос
	batch := &pgx.Batch{}
	for _, order := range orders {
		query := `
			INSERT INTO wb_orders (
				srid, g_number, date, last_change_date, supplier_article,
				tech_size, barcode, total_price, discount_percent, warehouse_name,is_cancel, dest_city_name, country_name, oblast_okrug_name,region_name, nm_id, category, brand
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

	// Отправляем пакет одним запросом
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	// Считаем количество обработанных строк
	var totalRows int64
	for i := 0; i < len(orders); i++ {
		ct, err := br.Exec()
		if err != nil {
			return 0, fmt.Errorf("ошибка вставки заказа %d: %w", i, err)
		}

		totalRows += ct.RowsAffected()
	}

	return int(totalRows), nil
}

// GetList возвращает список заказов с фильтрацией и пагинацией.
func (r *WbOrderRepo) GetList(ctx context.Context, filter domain.OrderFilter) ([]domain.WbOrder, int, error) {
	// Строим динамический WHERE
	where := "1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.From != "" {
		where += fmt.Sprintf(" AND date >= $%d", argIdx)
		args = append(args, filter.From)
		argIdx++
	}
	if filter.To != "" {
		where += fmt.Sprintf(" AND date <= $%d", argIdx)
		args = append(args, filter.To)
		argIdx++
	}

	// Запрос на получение записей
	query := fmt.Sprintf(`
		SELECT srid, g_number, date, last_change_date, supplier_article,
		       tech_size, barcode, total_price, discount_percent, warehouse_name, is_cancel, dest_city_name, country_name, oblast_okrug_name, region_name, nm_id, category, brand, created_at, updated_at
		FROM wb_orders
		WHERE %s
		ORDER BY date DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	var orders []domain.WbOrder
	if err := pgxscan.Select(ctx, r.pool, &orders, query, args...); err != nil {
		return nil, 0, fmt.Errorf("ошибка получения заказов: %w", err)
	}

	// Запрос на общее количество (для пагинации)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM wb_orders WHERE %s", where)
	var total int
	if err := pgxscan.Get(ctx, r.pool, &total, countQuery, args[:len(args)-2]...); err != nil {
		return nil, 0, fmt.Errorf("ошибка подсчёта заказов: %w", err)
	}

	return orders, total, nil
}

// GetStats возвращает агрегированную статистику по заказам WB.
func (r *WbOrderRepo) GetStats(ctx context.Context) (*domain.WbOrderStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_orders,
			SUM(CASE WHEN is_cancel THEN 1 ELSE 0 END) as cancelled_orders,
			SUM(total_price) as total_revenue,
			COUNT(DISTINCT nm_id) as unique_products
		FROM wb_orders
	`
	var stats domain.WbOrderStats
	if err := pgxscan.Get(ctx, r.pool, &stats, query); err != nil {
		return nil, fmt.Errorf("ошибка получения статистики: %w", err)
	}
	return &stats, nil
}

// CountForPeriod возвращает количество заказов по дням для графика.
func (r *WbOrderRepo) CountForPeriod(ctx context.Context, from, to string) ([]domain.DailyChartItem, error) {
	query := `
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
		return nil, fmt.Errorf("ошибка получения графика: %w", err)
	}
	return items, nil
}
