package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/atilatair/realput-bg/backend/internal/repository"
	"github.com/atilatair/realput-bg/backend/internal/testutil"
	"github.com/google/uuid"
)

func testAuthService(t *testing.T) *Service {
	t.Helper()
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)
	cfg := config.AuthConfig{
		SessionCookieName:  "realput_session",
		SessionTTL:         24 * time.Hour,
		GoogleClientID:     "test-id",
		GoogleClientSecret: "test-secret",
		GoogleRedirectURL:  "http://localhost:3000/callback",
		BcryptCost:         10,
	}
	return New(store, cfg)
}

func TestSignUpEmail_And_SignIn(t *testing.T) {
	svc := testAuthService(t)
	ctx := context.Background()
	email := fmt.Sprintf("auth-%s@example.com", uuid.NewString())

	user, token, err := svc.SignUpEmail(ctx, model.SignUpRequest{
		Email: email, Password: "password123", Name: "Auth Test",
	}, "127.0.0.1", "unit-test")
	if err != nil || user == nil || token == "" {
		t.Fatalf("signup: %v user=%#v token=%q", err, user, token)
	}

	_, _, err = svc.SignUpEmail(ctx, model.SignUpRequest{
		Email: email, Password: "password123", Name: "Dup",
	}, "127.0.0.1", "unit-test")
	if err != ErrEmailExists {
		t.Fatalf("expected duplicate email, got %v", err)
	}

	user2, token2, err := svc.SignInEmail(ctx, model.SignInRequest{
		Email: email, Password: "password123",
	}, "127.0.0.1", "unit-test")
	if err != nil || user2.Email != email || token2 == "" {
		t.Fatalf("signin: %v", err)
	}

	sess, err := svc.GetSession(ctx, token2)
	if err != nil || sess.User == nil || sess.User.Email != email {
		t.Fatalf("session: %v %#v", err, sess)
	}

	if err := svc.SignOut(ctx, token2); err != nil {
		t.Fatal(err)
	}
}

func TestSignUpValidation(t *testing.T) {
	svc := testAuthService(t)
	_, _, err := svc.SignUpEmail(context.Background(), model.SignUpRequest{
		Email: "", Password: "short",
	}, "", "")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSignInInvalidCredentials(t *testing.T) {
	svc := testAuthService(t)
	_, _, err := svc.SignInEmail(context.Background(), model.SignInRequest{
		Email: "nobody@example.com", Password: "wrongpass",
	}, "", "")
	if err != ErrInvalidCredentials {
		t.Fatalf("got %v", err)
	}
}

func TestUpdateProfile(t *testing.T) {
	svc := testAuthService(t)
	ctx := context.Background()
	email := fmt.Sprintf("profile-%s@example.com", uuid.NewString())

	user, token, err := svc.SignUpEmail(ctx, model.SignUpRequest{
		Email: email, Password: "password123", Name: "Before",
	}, "", "")
	if err != nil {
		t.Fatal(err)
	}
	sess, _ := svc.GetSession(ctx, token)
	_ = sess

	updated, err := svc.UpdateProfile(ctx, user.ID, "After Name")
	if err != nil || updated.Name != "After Name" {
		t.Fatalf("update: %v %#v", err, updated)
	}
}

func TestGoogleAuthURL(t *testing.T) {
	svc := testAuthService(t)
	url := svc.GoogleAuthURL("state123")
	if url == "" || !contains(url, "state=state123") {
		t.Fatalf("url %q", url)
	}
}

func TestRandomToken(t *testing.T) {
	a, err := randomToken(32)
	if err != nil || len(a) < 20 {
		t.Fatalf("token %q err %v", a, err)
	}
	b, _ := randomToken(32)
	if a == b {
		t.Fatal("expected unique tokens")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexSubstring(s, sub) >= 0)
}

func indexSubstring(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}