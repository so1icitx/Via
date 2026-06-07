package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/model"
)

func TestGetSession_EmptyToken(t *testing.T) {
	svc := testAuthService(t)
	sess, err := svc.GetSession(context.Background(), "")
	if err != nil || sess.User != nil {
		t.Fatalf("expected empty session, got %#v err=%v", sess, err)
	}
}

func TestSignOut_EmptyToken(t *testing.T) {
	svc := testAuthService(t)
	if err := svc.SignOut(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateProfile_EmptyName(t *testing.T) {
	svc := testAuthService(t)
	user, _, err := svc.SignUpEmail(context.Background(), model.SignUpRequest{
		Email: fmt.Sprintf("profile-empty-%d@example.com", time.Now().UnixNano()),
		Password: "password123", Name: "X",
	}, "", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.UpdateProfile(context.Background(), user.ID, "   ")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCookieConfig(t *testing.T) {
	svc := testAuthService(t)
	cfg := svc.CookieConfig()
	if cfg.SessionCookieName == "" {
		t.Fatal("expected cookie name")
	}
}