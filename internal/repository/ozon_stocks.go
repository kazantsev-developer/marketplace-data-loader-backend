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

// OzonRemainRepo is a PostgreSQL implementation of domain.OzonRemainRepository
type OzonRemainRepo struct {
	pool *pgxpool.Pool
}

// NewOzonRemainRepo returns a new OzonRemainRepo instance
func NewOzonRemainRepo(pool *pgxpool.Pool) *OzonRemainRepo {
	return &OzonRemainRepo{pool: pool}
}

// UpsertBatch inserts or updates a batch of Ozon stock records
func (r *OzonRemainRepo) UpsertBatch(ctx context.Context, remains []domain.OzonRemain) (int, error) {
	if len(remains) == 0 {
		return 0, nil
	}

	const query = `
		INSERT INTO ozon_remains (
			sku, product_id, item_code, category, brand, name,
			fbo_visible_amount, fbo_present_amount, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
		ON CONFLICT (sku) DO UPDATE SET
			product_id = EXCLUDED.product_id,
			item_code = EXCLUDED.item_code,
			category = EXCLUDED.category,
			brand = EXCLUDED.brand,
			name = EXCLUDED.name,
			fbo_visible_amount = EXCLUDED.fbo_visible_amount,
			fbo_present_amount = EXCLUDED.fbo_present_amount,
			synced_at = CURRENT_TIMESTAMP
	`

	batch := &pgx.Batch{}
	for _, remain := range remains {
		batch.Queue(query,
			remain.Sku,
			remain.ProductID,
			remain.ItemCode,
			remain.Category,
			remain.Brand,
			remain.Name,
			remain.FboVisibleAmount,
			remain.FboPresentAmount,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	var totalRows int64
	for i := range remains {
		ct, err := br.Exec()
		if err != nil {
			return 0, fmt.Errorf("execute batch item %d: %w", i, err)
		}
		totalRows += ct.RowsAffected()
	}

	return int(totalRows), nil
}

// GetAll retrieves Ozon stock records filtered by brand or search term
func (r *OzonRemainRepo) GetAll(ctx context.Context, brand, search string) ([]domain.OzonRemain, error) {
	var (
		sb   strings.Builder
		args = make([]any, 0, 2)
		idx  = 1
	)

	sb.Grow(256)
	sb.WriteString(`SELECT sku, product_id, item_code, category, brand, name, fbo_visible_amount, fbo_present_amount, updated_at, synced_at FROM ozon_remains WHERE 1=1`)

	if brand != "" {
		fmt.Fprintf(&sb, " AND brand = $%d", idx)
		args = append(args, brand)
		idx++
	}
	if search != "" {
		pattern := "%" + search + "%"
		fmt.Fprintf(&sb, " AND (item_code ILIKE $%d OR name ILIKE $%d)", idx, idx+1)
		args = append(args, pattern, pattern)
		idx += 2
	}

	sb.WriteString(" ORDER BY fbo_visible_amount DESC")

	var remains []domain.OzonRemain
	if err := pgxscan.Select(ctx, r.pool, &remains, sb.String(), args...); err != nil {
		return nil, fmt.Errorf("select ozon remains: %w", err)
	}

	return remains, nil
}

// ResetStale sets visible and present amounts to zero for products not synced after the given time
func (r *OzonRemainRepo) ResetStale(ctx context.Context, before string) (int, error) {
	const query = `
		UPDATE ozon_remains 
		SET fbo_visible_amount = 0, fbo_present_amount = 0, synced_at = CURRENT_TIMESTAMP
		WHERE synced_at < $1
	`

	ct, err := r.pool.Exec(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("reset stale ozon remains: %w", err)
	}
	return int(ct.RowsAffected()), nil
}

// GetStats returns aggregated statistics about Ozon stock
func (r *OzonRemainRepo) GetStats(ctx context.Context) (*domain.OzonRemainStats, error) {
	const totalQuery = `SELECT COUNT(*) FROM ozon_remains`
	const amountQuery = `SELECT SUM(fbo_visible_amount) as total_visible, SUM(fbo_present_amount) as total_present FROM ozon_remains`
	const brandQuery = `
		SELECT brand, COUNT(*) as products, SUM(fbo_visible_amount) as visible 
		FROM ozon_remains WHERE brand IS NOT NULL 
		GROUP BY brand ORDER BY visible DESC LIMIT 10`
	const updatedQuery = `SELECT COUNT(*) FROM ozon_remains WHERE updated_at > NOW() - INTERVAL '1 hour'`

	var total int
	if err := pgxscan.Get(ctx, r.pool, &total, totalQuery); err != nil {
		return nil, fmt.Errorf("get total ozon remains: %w", err)
	}

	var visible, present int
	row := r.pool.QueryRow(ctx, amountQuery)
	if err := row.Scan(&visible, &present); err != nil {
		return nil, fmt.Errorf("get ozon remain amounts: %w", err)
	}

	rows, err := r.pool.Query(ctx, brandQuery)
	if err != nil {
		return nil, fmt.Errorf("query ozon brands: %w", err)
	}
	defer rows.Close()

	brands := make([]domain.BrandStat, 0, 10)
	for rows.Next() {
		var b domain.BrandStat
		if err := rows.Scan(&b.Brand, &b.Products, &b.Visible); err != nil {
			return nil, fmt.Errorf("scan brand row: %w", err)
		}
		brands = append(brands, b)
	}

	var updated int
	if err := pgxscan.Get(ctx, r.pool, &updated, updatedQuery); err != nil {
		return nil, fmt.Errorf("get updated ozon remains: %w", err)
	}

	return &domain.OzonRemainStats{
		TotalProducts:   total,
		TotalVisible:    visible,
		TotalPresent:    present,
		TopBrands:       brands,
		UpdatedLastHour: updated,
	}, nil
}
