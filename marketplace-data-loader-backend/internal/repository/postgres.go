// Package repository provides implementations of domain repository interfaces using PostgreSQL.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kazantsev-developer/marketplace-data-loader-backend/internal/config"
)

func NewPool(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать DSN: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.PoolMax)
	poolCfg.MinConns = 2
	poolCfg.MaxConnIdleTime = time.Duration(cfg.PoolIdleTimeoutMs) * time.Millisecond
	poolCfg.ConnConfig.ConnectTimeout = time.Duration(cfg.PoolConnTimeoutMs) * time.Millisecond

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пул соединений: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	return pool, nil
}
