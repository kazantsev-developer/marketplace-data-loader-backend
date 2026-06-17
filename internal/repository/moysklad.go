// Package repository implements domain repository interfaces using PostgreSQL
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/domain"
)

// MoyskladRepo is a PostgreSQL implementation of domain.MoyskladRepository
type MoyskladRepo struct {
	pool *pgxpool.Pool
}

// NewMoyskladRepo returns a new MoyskladRepo instance
func NewMoyskladRepo(pool *pgxpool.Pool) *MoyskladRepo {
	return &MoyskladRepo{pool: pool}
}

// UpsertStores inserts or updates a batch of MoySklad stores
func (r *MoyskladRepo) UpsertStores(ctx context.Context, stores []domain.MsStore) (int, error) {
	if len(stores) == 0 {
		return 0, nil
	}

	const query = `
		INSERT INTO ms_stores (uuid, name, code, external_code, address, created_at, updated_at, synced_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (uuid) DO UPDATE SET
			name = EXCLUDED.name,
			code = EXCLUDED.code,
			external_code = EXCLUDED.external_code,
			address = EXCLUDED.address,
			updated_at = EXCLUDED.updated_at,
			synced_at = CURRENT_TIMESTAMP
	`

	batch := &pgx.Batch{}
	for _, store := range stores {
		batch.Queue(query,
			store.UUID,
			store.Name,
			store.Code,
			store.ExternalCode,
			store.Address,
			store.CreatedAt,
			store.UpdatedAt,
			time.Now(),
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	var totalRows int64
	for range stores {
		ct, err := br.Exec()
		if err != nil {
			return 0, fmt.Errorf("execute batch item: %w", err)
		}
		totalRows += ct.RowsAffected()
	}

	return int(totalRows), nil
}

// GetStores returns all MoySklad stores ordered by name
func (r *MoyskladRepo) GetStores(ctx context.Context) ([]domain.MsStore, error) {
	const query = `SELECT uuid, name, code, external_code, address, created_at, updated_at, synced_at FROM ms_stores ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query stores: %w", err)
	}
	defer rows.Close()

	stores := make([]domain.MsStore, 0)
	for rows.Next() {
		var s domain.MsStore
		if err := rows.Scan(
			&s.UUID, &s.Name, &s.Code, &s.ExternalCode, &s.Address,
			&s.CreatedAt, &s.UpdatedAt, &s.SyncedAt,
		); err != nil {
			return nil, fmt.Errorf("scan store: %w", err)
		}
		stores = append(stores, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return stores, nil
}

// CreateSnapshot inserts a new snapshot and returns its ID
func (r *MoyskladRepo) CreateSnapshot(ctx context.Context) (int, error) {
	const query = `
		INSERT INTO ms_snapshots (collected_at)
		VALUES (CURRENT_TIMESTAMP)
		RETURNING id
	`

	var id int
	if err := r.pool.QueryRow(ctx, query).Scan(&id); err != nil {
		return 0, fmt.Errorf("create snapshot: %w", err)
	}
	return id, nil
}

// InsertStockDetails inserts a batch of stock detail records using pgx.Batch
func (r *MoyskladRepo) InsertStockDetails(ctx context.Context, details []domain.MsStockDetail) (int, error) {
	if len(details) == 0 {
		return 0, nil
	}

	const query = `
		INSERT INTO ms_stock_details (snapshot_id, product_uuid, store_uuid, stock, reserve, in_transit)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (snapshot_id, product_uuid, store_uuid) DO UPDATE SET
			stock = EXCLUDED.stock,
			reserve = EXCLUDED.reserve,
			in_transit = EXCLUDED.in_transit
	`

	batch := &pgx.Batch{}
	for _, d := range details {
		batch.Queue(query,
			d.SnapshotID,
			d.ProductUUID,
			d.StoreUUID,
			d.Stock,
			d.Reserve,
			d.InTransit,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	var totalRows int64
	for range details {
		ct, err := br.Exec()
		if err != nil {
			return 0, fmt.Errorf("execute batch item: %w", err)
		}
		totalRows += ct.RowsAffected()
	}

	return int(totalRows), nil
}

// UpsertProductTotals inserts or updates a batch of product stock totals
func (r *MoyskladRepo) UpsertProductTotals(ctx context.Context, totals []domain.MsProductTotal) (int, error) {
	if len(totals) == 0 {
		return 0, nil
	}

	const query = `
		INSERT INTO ms_product_totals (product_uuid, article, name, total_stock, total_reserve, total_in_transit, snapshot_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (product_uuid) DO UPDATE SET
			article = EXCLUDED.article,
			name = EXCLUDED.name,
			total_stock = EXCLUDED.total_stock,
			total_reserve = EXCLUDED.total_reserve,
			total_in_transit = EXCLUDED.total_in_transit,
			snapshot_id = EXCLUDED.snapshot_id,
			updated_at = CURRENT_TIMESTAMP
	`

	batch := &pgx.Batch{}
	for _, t := range totals {
		batch.Queue(query,
			t.ProductUUID,
			t.Article,
			t.Name,
			t.TotalStock,
			t.TotalReserve,
			t.TotalInTransit,
			t.SnapshotID,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	var totalRows int64
	for range totals {
		ct, err := br.Exec()
		if err != nil {
			return 0, fmt.Errorf("execute batch item: %w", err)
		}
		totalRows += ct.RowsAffected()
	}

	return int(totalRows), nil
}

// GetStockDetails returns stock details filtered by product and/or store UUID
func (r *MoyskladRepo) GetStockDetails(ctx context.Context, productUUID, storeUUID string) ([]domain.MsStockDetail, error) {
	query := `
		SELECT msd.snapshot_id, msd.product_uuid, msd.store_uuid, msd.stock, msd.reserve, msd.in_transit, msd.created_at
		FROM ms_stock_details msd
		JOIN ms_product_totals p ON msd.product_uuid = p.product_uuid
		JOIN ms_stores s ON msd.store_uuid = s.uuid
		WHERE 1=1
	`
	args := make([]any, 0, 2)
	idx := 1

	if productUUID != "" {
		query += fmt.Sprintf(" AND msd.product_uuid = $%d", idx)
		args = append(args, productUUID)
		idx++
	}
	if storeUUID != "" {
		query += fmt.Sprintf(" AND msd.store_uuid = $%d", idx)
		args = append(args, storeUUID)
		idx++
	}
	query += " ORDER BY msd.snapshot_id DESC LIMIT 1000"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query stock details: %w", err)
	}
	defer rows.Close()

	details := make([]domain.MsStockDetail, 0)
	for rows.Next() {
		var d domain.MsStockDetail
		if err := rows.Scan(&d.SnapshotID, &d.ProductUUID, &d.StoreUUID, &d.Stock, &d.Reserve, &d.InTransit, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan stock detail: %w", err)
		}
		details = append(details, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return details, nil
}

// GetProductTotals returns all product aggregates ordered by total_stock descending
func (r *MoyskladRepo) GetProductTotals(ctx context.Context) ([]domain.MsProductTotal, error) {
	const query = `
		SELECT product_uuid, article, name, total_stock, total_reserve, total_in_transit, snapshot_id, updated_at, created_at
		FROM ms_product_totals
		ORDER BY total_stock DESC
		LIMIT 100
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query product totals: %w", err)
	}
	defer rows.Close()

	totals := make([]domain.MsProductTotal, 0)
	for rows.Next() {
		var t domain.MsProductTotal
		if err := rows.Scan(
			&t.ProductUUID, &t.Article, &t.Name,
			&t.TotalStock, &t.TotalReserve, &t.TotalInTransit,
			&t.SnapshotID, &t.UpdatedAt, &t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan product total: %w", err)
		}
		totals = append(totals, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return totals, nil
}

// GetStockTotal returns the sum of all stock across all products
func (r *MoyskladRepo) GetStockTotal(ctx context.Context) (int, error) {
	const query = `SELECT COALESCE(SUM(total_stock), 0) FROM ms_product_totals`

	var total int
	if err := r.pool.QueryRow(ctx, query).Scan(&total); err != nil {
		return 0, fmt.Errorf("query stock total: %w", err)
	}
	return total, nil
}
