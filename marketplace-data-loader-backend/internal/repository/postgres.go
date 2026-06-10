// Package repository provides implementations of domain repository interfaces using PostgreSQL
package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/config"
)

// NewPool creates and initializes a new PostgreSQL connection pool
func NewPool(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {
	safePassword := strings.ReplaceAll(cfg.Password, "'", "\\'")
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password='%s' dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		safePassword,
		cfg.Name,
		cfg.SSLMode,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn config: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.PoolMax)
	poolCfg.MinConns = 2
	poolCfg.MaxConnIdleTime = time.Duration(cfg.PoolIdleTimeoutMs) * time.Millisecond
	poolCfg.ConnConfig.ConnectTimeout = time.Duration(cfg.PoolConnTimeoutMs) * time.Millisecond

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, poolCfg.ConnConfig.ConnectTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}
