package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/atilatair/realput-bg/backend/internal/testutil"
	"github.com/google/uuid"
)

func TestIntegration_UserSessionLifecycle(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := New(pool)
	ctx := context.Background()
	email := fmt.Sprintf("test-%s@example.com", uuid.NewString())

	user, err := store.CreateUser(ctx, "Test User", email, false)
	if err != nil {
		t.Fatal(err)
	}

	hash := "$2a$12$abcdefghijklmnopqrstuv/abcdefghijklmnopqrstuvabcdefghi"
	if err := store.CreateCredentialAccount(ctx, user.ID, hash); err != nil {
		t.Fatal(err)
	}

	gotHash, err := store.GetCredentialPasswordHash(ctx, user.ID)
	if err != nil || gotHash != hash {
		t.Fatalf("hash mismatch: %v %q", err, gotHash)
	}

	byEmail, err := store.FindUserByEmail(ctx, email)
	if err != nil || byEmail == nil || byEmail.ID != user.ID {
		t.Fatalf("find by email failed: %v %#v", err, byEmail)
	}

	byID, err := store.FindUserByID(ctx, user.ID)
	if err != nil || byID == nil || byID.Email != email {
		t.Fatalf("find by id failed: %v %#v", err, byID)
	}

	if err := store.UpdateUserName(ctx, user.ID, "Updated Name"); err != nil {
		t.Fatal(err)
	}
	updated, err := store.FindUserByID(ctx, user.ID)
	if err != nil || updated.Name != "Updated Name" {
		t.Fatalf("update name failed: %v %#v", err, updated)
	}

	token := "sess-" + uuid.NewString()
	sess, err := store.CreateSession(ctx, user.ID, token, "127.0.0.1", "test", time.Now().UTC().Add(time.Hour))
	if err != nil || sess.Token != token {
		t.Fatalf("create session: %v %#v", err, sess)
	}

	found, err := store.FindSessionByToken(ctx, token)
	if err != nil || found == nil || found.UserID != user.ID {
		t.Fatalf("find session: %v %#v", err, found)
	}

	if err := store.TouchSession(ctx, found.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteSessionByToken(ctx, token); err != nil {
		t.Fatal(err)
	}
}

func TestIntegration_SavedResults(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := New(pool)
	ctx := context.Background()
	email := fmt.Sprintf("saved-%s@example.com", uuid.NewString())

	user, err := store.CreateUser(ctx, "Saver", email, true)
	if err != nil {
		t.Fatal(err)
	}

	payload := model.ResultsPayload{
		Intro: "saved",
		Items: []model.ResultItem{{ID: "p1", Title: "Uni", Summary: "s"}},
	}
	id, err := store.InsertSavedResult(ctx, user.ID, "query text", payload)
	if err != nil || id == 0 {
		t.Fatalf("create saved: %v id=%d", err, id)
	}


	rows, err := store.ListSavedResults(ctx, user.ID)
	if err != nil || len(rows) == 0 {
		t.Fatalf("list saved: %v len=%d", err, len(rows))
	}
	if err := store.DeleteSavedResult(ctx, user.ID, id); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteSavedResult(ctx, user.ID, 999999); err == nil {
		t.Fatal("expected error when deleting missing row")
	}
}

func TestIntegration_GoogleAccount(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := New(pool)
	ctx := context.Background()
	email := fmt.Sprintf("google-%s@example.com", uuid.NewString())
	googleID := "google-" + uuid.NewString()

	user, err := store.CreateUser(ctx, "Google User", email, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateGoogleAccount(ctx, user.ID, googleID); err != nil {
		t.Fatal(err)
	}
	found, err := store.FindGoogleAccount(ctx, googleID)
	if err != nil || found == nil || found.ID != user.ID {
		t.Fatalf("find google account: %v %#v", err, found)
	}
}