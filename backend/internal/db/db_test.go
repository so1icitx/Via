package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/config"
)

func testDatabaseURL() string {
	if v := os.Getenv("TEST_DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://realput:realput@localhost:5432/realput?sslmode=disable"
}

func TestConnect_Postgres(t *testing.T) {
	pool, err := Connect(context.Background(), config.DatabaseConfig{
		URL:             testDatabaseURL(),
		MaxConns:        4,
		MinConns:        1,
		MaxConnLifetime: time.Minute,
		MigrateOnStart:  true,
	})
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	defer pool.Close()

	if pool.Pool == nil {
		t.Fatal("nil pool")
	}
}

func TestConnect_InvalidURL(t *testing.T) {
	_, err := Connect(context.Background(), config.DatabaseConfig{URL: "not-a-url"})
	if err == nil {
		t.Fatal("expected error")
	}
}