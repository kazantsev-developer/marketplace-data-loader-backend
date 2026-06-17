// Package repository implements domain repository interfaces using PostgreSQL
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/domain"
)

// SyncLogRepo is a PostgreSQL implementation of domain.SyncLogRepository
type SyncLogRepo struct {
	pool *pgxpool.Pool
}

// NewSyncLogRepo returns a new SyncLogRepo instance
func NewSyncLogRepo(pool *pgxpool.Pool) *SyncLogRepo {
	return &SyncLogRepo{pool: pool}
}

// Insert saves a new synchronization log entry and returns its ID
func (r *SyncLogRepo) Insert(ctx context.Context, log domain.SyncLog) (int, error) {
	const query = `
		INSERT INTO sync_logs (
			sync_at, status, records_count, date_from, date_to,
			error_message, pages_count, execution_time_seconds, entity_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var id int
	err := r.pool.QueryRow(ctx, query,
		log.SyncAt,
		log.Status,
		log.RecordsCount,
		log.DateFrom,
		log.DateTo,
		log.ErrorMessage,
		log.PagesCount,
		log.ExecutionTimeSeconds,
		log.EntityType,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert sync log: %w", err)
	}
	return id, nil
}

// GetList returns recent sync logs for a given entity type
func (r *SyncLogRepo) GetList(ctx context.Context, entityType string, limit int) ([]domain.SyncLog, error) {
	query := `
		SELECT id, sync_at, status, records_count, date_from, date_to,
		       error_message, pages_count, execution_time_seconds, entity_type
		FROM sync_logs
		WHERE entity_type = $1
		ORDER BY sync_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, entityType, limit)
	if err != nil {
		return nil, fmt.Errorf("query sync logs: %w", err)
	}
	defer rows.Close()

	logs := make([]domain.SyncLog, 0, limit)
	for rows.Next() {
		var l domain.SyncLog
		if err := rows.Scan(
			&l.ID, &l.SyncAt, &l.Status, &l.RecordsCount,
			&l.DateFrom, &l.DateTo, &l.ErrorMessage,
			&l.PagesCount, &l.ExecutionTimeSeconds, &l.EntityType,
		); err != nil {
			return nil, fmt.Errorf("scan sync log: %w", err)
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return logs, nil
}

// InsertOzonLog saves an Ozon synchronization log entry and returns its ID
func (r *SyncLogRepo) InsertOzonLog(ctx context.Context, log domain.OzonSyncLog) (int, error) {
	const query = `
		INSERT INTO ozon_sync_logs (status, scheme, records_count, date_from, error_message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int
	err := r.pool.QueryRow(ctx, query,
		log.Status,
		log.Scheme,
		log.RecordsCount,
		log.DateFrom,
		log.ErrorMessage,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert ozon sync log: %w", err)
	}
	return id, nil
}

// InsertMsJobLog saves a MoySklad job log entry and returns its ID
func (r *SyncLogRepo) InsertMsJobLog(ctx context.Context, log domain.MsJobLog) (int, error) {
	const query = `
		INSERT INTO ms_job_log (
			status, records_count, details_count, aggregates_count,
			error_message, execution_time_seconds
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int
	err := r.pool.QueryRow(ctx, query,
		log.Status,
		log.RecordsCount,
		log.DetailsCount,
		log.AggregatesCount,
		log.ErrorMessage,
		log.ExecutionTimeSeconds,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert ms job log: %w", err)
	}
	return id, nil
}
