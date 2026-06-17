// Package repository implements domain repository interfaces using PostgreSQL
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/domain"
)

// WbCardRepo is a PostgreSQL implementation of domain.WbCardRepository
type WbCardRepo struct {
	pool *pgxpool.Pool
}

// NewWbCardRepo returns a new WbCardRepo instance
func NewWbCardRepo(pool *pgxpool.Pool) *WbCardRepo {
	return &WbCardRepo{pool: pool}
}

// UpsertBatch inserts or updates a batch of WB product cards
func (r *WbCardRepo) UpsertBatch(ctx context.Context, cards []domain.WbCard) (int, error) {
	if len(cards) == 0 {
		return 0, nil
	}

	const query = `
		INSERT INTO wb_cards (
			nm_id, vendor_code, brand, title, description,
			category, subject, characteristics, sizes, photos,
			video, dimensions, weight, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14
		)
		ON CONFLICT (nm_id) DO UPDATE SET
			vendor_code = EXCLUDED.vendor_code,
			brand = EXCLUDED.brand,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			category = EXCLUDED.category,
			subject = EXCLUDED.subject,
			characteristics = EXCLUDED.characteristics,
			sizes = EXCLUDED.sizes,
			photos = EXCLUDED.photos,
			video = EXCLUDED.video,
			dimensions = EXCLUDED.dimensions,
			weight = EXCLUDED.weight,
			updated_at = EXCLUDED.updated_at,
			synced_at = CURRENT_TIMESTAMP
	`

	batch := &pgx.Batch{}
	for _, card := range cards {
		batch.Queue(query,
			card.NmID,
			card.VendorCode,
			card.Brand,
			card.Title,
			card.Description,
			card.Category,
			card.Subject,
			card.Characteristics,
			card.Sizes,
			card.Photos,
			card.Video,
			card.Dimensions,
			card.Weight,
			card.UpdatedAt,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	var totalRows int64
	for i := range cards {
		ct, err := br.Exec()
		if err != nil {
			return 0, fmt.Errorf("execute batch item %d: %w", i, err)
		}
		totalRows += ct.RowsAffected()
	}

	return int(totalRows), nil
}

// GetList returns a paginated list of WB product cards, optionally filtered by search term
func (r *WbCardRepo) GetList(ctx context.Context, search string, limit, offset int) ([]domain.WbCard, int, error) {
	var (
		sb   strings.Builder
		args = make([]any, 0, 5)
		idx  = 1
	)

	sb.Grow(512)
	sb.WriteString(`WHERE 1=1`)

	if search != "" {
		pattern := "%" + search + "%"
		fmt.Fprintf(&sb, " AND (vendor_code ILIKE $%d OR title ILIKE $%d OR brand ILIKE $%d)", idx, idx+1, idx+2)
		args = append(args, pattern, pattern, pattern)
		idx += 3
	}

	whereClause := sb.String()

	sb.Reset()
	sb.WriteString(`SELECT nm_id, vendor_code, brand, title, description, category, subject, characteristics, sizes, photos, video, dimensions, weight, updated_at, created_at, synced_at FROM wb_cards `)
	sb.WriteString(whereClause)
	sb.WriteString(" ORDER BY updated_at DESC")

	fmt.Fprintf(&sb, " LIMIT $%d OFFSET $%d", idx, idx+1)
	dataArgs := append(args, limit, offset)

	var cards []domain.WbCard
	if err := pgxscan.Select(ctx, r.pool, &cards, sb.String(), dataArgs...); err != nil {
		return nil, 0, fmt.Errorf("select cards: %w", err)
	}

	sb.Reset()
	sb.WriteString("SELECT COUNT(*) FROM wb_cards ")
	sb.WriteString(whereClause)

	var total int
	if err := pgxscan.Get(ctx, r.pool, &total, sb.String(), args...); err != nil {
		return nil, 0, fmt.Errorf("count cards: %w", err)
	}

	return cards, total, nil
}

// GetStats returns aggregated statistics about WB product cards
func (r *WbCardRepo) GetStats(ctx context.Context) (*domain.WbCardStats, error) {
	const query = `
		SELECT
			COUNT(*) AS total_cards,
			COUNT(*) FILTER (WHERE synced_at > NOW() - INTERVAL '1 hour') AS updated_last_hour
		FROM wb_cards
	`

	var stats domain.WbCardStats
	if err := pgxscan.Get(ctx, r.pool, &stats, query); err != nil {
		return nil, fmt.Errorf("get card stats: %w", err)
	}

	return &stats, nil
}

// GetCursor returns the last synchronization cursor state
func (r *WbCardRepo) GetCursor(ctx context.Context) (*domain.SyncCursorState, error) {
	const query = `SELECT last_updated_at, last_nm_id FROM sync_cursor_state WHERE id = 1`

	var cursor domain.SyncCursorState
	if err := pgxscan.Get(ctx, r.pool, &cursor, query); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get cursor: %w", err)
	}

	return &cursor, nil
}

// SaveCursor updates the pagination cursor for incremental card synchronization
func (r *WbCardRepo) SaveCursor(ctx context.Context, updatedAt string, nmID int64) error {
	const query = `
		INSERT INTO sync_cursor_state (id, last_updated_at, last_nm_id)
		VALUES (1, $1, $2)
		ON CONFLICT (id) DO UPDATE SET
			last_updated_at = EXCLUDED.last_updated_at,
			last_nm_id = EXCLUDED.last_nm_id,
			updated_at = CURRENT_TIMESTAMP
	`

	if _, err := r.pool.Exec(ctx, query, updatedAt, nmID); err != nil {
		return fmt.Errorf("upsert cursor: %w", err)
	}

	return nil
}
