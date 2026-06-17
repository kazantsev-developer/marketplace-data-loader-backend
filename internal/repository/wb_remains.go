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

// WbRemainRepo is a PostgreSQL implementation of domain.WbRemainRepository
type WbRemainRepo struct {
	pool *pgxpool.Pool
}

// NewWbRemainRepo returns a new WbRemainRepo instance
func NewWbRemainRepo(pool *pgxpool.Pool) *WbRemainRepo {
	return &WbRemainRepo{pool: pool}
}

// UpsertBatch inserts or updates a batch of WB remains using a single round-trip
func (r *WbRemainRepo) UpsertBatch(ctx context.Context, remains []domain.WbRemain) (int, error) {
	if len(remains) == 0 {
		return 0, nil
	}

	const query = `
		INSERT INTO wb_remains (nm_id, size, warehouse, quantity, barcode)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (nm_id, warehouse, size) DO UPDATE SET
			quantity = EXCLUDED.quantity,
			barcode = EXCLUDED.barcode,
			updated_at = CURRENT_TIMESTAMP
	`

	batch := &pgx.Batch{}
	for _, remain := range remains {
		batch.Queue(query,
			remain.NmID,
			remain.Size,
			remain.Warehouse,
			remain.Quantity,
			remain.Barcode,
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

// GetAll retrieves all WB remains, optionally filtered by warehouse or search term
func (r *WbRemainRepo) GetAll(ctx context.Context, warehouse, search string) ([]domain.WbRemain, error) {
	var (
		sb   strings.Builder
		args = make([]any, 0, 3)
		idx  = 1
	)

	sb.Grow(256)
	sb.WriteString(`SELECT nm_id, size, warehouse, quantity, barcode, updated_at FROM wb_remains WHERE 1=1`)

	if warehouse != "" {
		fmt.Fprintf(&sb, " AND warehouse = $%d", idx)
		args = append(args, warehouse)
		idx++
	}

	if search != "" {
		pattern := "%" + search + "%"
		fmt.Fprintf(&sb, " AND (nm_id::text ILIKE $%d OR barcode ILIKE $%d)", idx, idx+1)
		args = append(args, pattern, pattern)
		idx += 2
	}

	sb.WriteString(" ORDER BY warehouse, nm_id")

	var remains []domain.WbRemain
	if err := pgxscan.Select(ctx, r.pool, &remains, sb.String(), args...); err != nil {
		return nil, fmt.Errorf("select remains: %w", err)
	}

	return remains, nil
}
