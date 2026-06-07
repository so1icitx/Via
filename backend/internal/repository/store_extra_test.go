package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/atilatair/realput-bg/backend/internal/testutil"
	"github.com/google/uuid"
)

func TestFindUserByID_NotFound(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := New(pool)
	user, err := store.FindUserByEmail(context.Background(), fmt.Sprintf("missing-%s@example.com", uuid.NewString()))
	if err != nil || user != nil {
		t.Fatalf("err=%v user=%#v", err, user)
	}
}

func TestGetCredentialPasswordHash_Missing(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := New(pool)
	hash, err := store.GetCredentialPasswordHash(context.Background(), uuid.NewString())
	if err != nil || hash != "" {
		t.Fatalf("err=%v hash=%q", err, hash)
	}
}