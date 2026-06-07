package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresURL returns the integration test database URL.
func PostgresURL() string {
	if v := os.Getenv("TEST_DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://realput:realput@localhost:5432/realput?sslmode=disable"
}

// OpenPostgres connects to Postgres or skips the test when unavailable.
func OpenPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, PostgresURL())
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("postgres ping failed: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}