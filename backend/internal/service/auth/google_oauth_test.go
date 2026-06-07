package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/atilatair/realput-bg/backend/internal/repository"
	"github.com/atilatair/realput-bg/backend/internal/testutil"
	"github.com/google/uuid"
)

func TestCompleteGoogleOAuth_NewUser(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)
	srv, profile := newOAuthMockServer(t, nil, 0, 0)
	defer srv.Close()

	svc := oauthTestService(t, store, srv)
	user, token, err := svc.CompleteGoogleOAuth(context.Background(), "auth-code", "127.0.0.1", "test")
	if err != nil || user == nil || token == "" {
		t.Fatalf("oauth failed: %v user=%#v token=%q", err, user, token)
	}
	if user.Email != profile.Email {
		t.Fatalf("email %q want %q", user.Email, profile.Email)
	}
}

func TestCompleteGoogleOAuth_ExistingGoogleAccount(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)
	ctx := context.Background()

	profile := &oauthMockProfile{ID: "gid-existing-" + uuid.NewString(), Email: fmt.Sprintf("existing-%s@example.com", uuid.NewString())}
	user, err := store.CreateUser(ctx, profile.Name, profile.Email, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateGoogleAccount(ctx, user.ID, profile.ID); err != nil {
		t.Fatal(err)
	}

	srv, _ := newOAuthMockServer(t, profile, 0, 0)
	defer srv.Close()
	svc := oauthTestService(t, store, srv)

	got, token, err := svc.CompleteGoogleOAuth(ctx, "code", "127.0.0.1", "test")
	if err != nil || got.ID != user.ID || token == "" {
		t.Fatalf("existing google: %v %#v token=%q", err, got, token)
	}
}

func TestCompleteGoogleOAuth_LinksExistingEmail(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)
	ctx := context.Background()

	email := fmt.Sprintf("link-%s@example.com", uuid.NewString())
	user, err := store.CreateUser(ctx, "Email User", email, false)
	if err != nil {
		t.Fatal(err)
	}

	profile := &oauthMockProfile{ID: "gid-link-" + uuid.NewString(), Email: email}
	srv, _ := newOAuthMockServer(t, profile, 0, 0)
	defer srv.Close()
	svc := oauthTestService(t, store, srv)

	got, token, err := svc.CompleteGoogleOAuth(ctx, "code", "127.0.0.1", "test")
	if err != nil || got.ID != user.ID || token == "" {
		t.Fatalf("link email: %v %#v token=%q", err, got, token)
	}
}

func TestCompleteGoogleOAuth_ExchangeError(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)
	srv, _ := newOAuthMockServer(t, nil, http.StatusBadRequest, 0)
	defer srv.Close()
	svc := oauthTestService(t, store, srv)

	_, _, err := svc.CompleteGoogleOAuth(context.Background(), "bad-code", "127.0.0.1", "test")
	if err == nil {
		t.Fatal("expected exchange error")
	}
}

func TestCompleteGoogleOAuth_ProfileError(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)
	srv, _ := newOAuthMockServer(t, nil, 0, http.StatusUnauthorized)
	defer srv.Close()
	svc := oauthTestService(t, store, srv)

	_, _, err := svc.CompleteGoogleOAuth(context.Background(), "code", "127.0.0.1", "test")
	if err == nil {
		t.Fatal("expected profile error")
	}
}

func TestCompleteGoogleOAuth_IncompleteProfile(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok","token_type":"bearer"}`))
		case "/userinfo":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "", "email": ""})
		}
	}))
	defer srv.Close()
	svc := oauthTestService(t, store, srv)

	_, _, err := svc.CompleteGoogleOAuth(context.Background(), "code", "127.0.0.1", "test")
	if err == nil {
		t.Fatal("expected incomplete profile error")
	}
}

func TestGetSession_Expired(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)
	ctx := context.Background()

	svc := testAuthService(t)
	email := fmt.Sprintf("expired-%s@example.com", uuid.NewString())
	user, _, err := svc.SignUpEmail(ctx, model.SignUpRequest{
		Email: email, Password: "password123", Name: "Expired",
	}, "", "")
	if err != nil {
		t.Fatal(err)
	}

	token := "expired-" + uuid.NewString()
	_, err = store.CreateSession(ctx, user.ID, token, "127.0.0.1", "test", time.Now().UTC().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	sess, err := svc.GetSession(ctx, token)
	if err != nil || sess.User != nil {
		t.Fatalf("expected empty expired session: %#v err=%v", sess, err)
	}
}