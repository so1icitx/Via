// Package db manages the PostgreSQL connection pool and schema migrations.
package db

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

// Pool wraps a pgx connection pool.
type Pool struct {
	*pgxpool.Pool
}

// Connect creates a pool and optionally applies embedded migrations.
func Connect(ctx context.Context, cfg config.DatabaseConfig) (*Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	ctxPing, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if cfg.MigrateOnStart {
		if _, err := pool.Exec(ctx, schemaSQL); err != nil {
			pool.Close()
			return nil, fmt.Errorf("apply migrations: %w", err)
		}
	}

	return &Pool{Pool: pool}, nil
}