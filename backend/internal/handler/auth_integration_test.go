package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/middleware"
	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/atilatair/realput-bg/backend/internal/repository"
	"github.com/atilatair/realput-bg/backend/internal/service/auth"
	"github.com/atilatair/realput-bg/backend/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func newAuthRouter(t *testing.T) (*gin.Engine, *auth.Service) {
	t.Helper()
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)
	cfg := config.AuthConfig{
		BaseURL:            "http://localhost:8080",
		FrontendURL:        "http://localhost:3000",
		SessionCookieName:  "realput_session",
		SessionTTL:         24 * time.Hour,
		CookieSecure:       false,
		CookieSameSite:     "lax",
		GoogleClientID:     "test-client",
		GoogleClientSecret: "test-secret",
		GoogleRedirectURL:  "http://localhost:3000/api/auth/callback/google",
		BcryptCost:         10,
	}
	svc := auth.New(store, cfg)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := NewAuthHandler(svc, logger)

	r := gin.New()
	r.POST("/api/auth/sign-up/email", h.SignUpEmail)
	r.POST("/api/auth/sign-in/email", h.SignInEmail)
	r.GET("/api/auth/get-session", h.GetSession)
	r.POST("/api/auth/sign-out", h.SignOut)
	r.PATCH("/api/auth/profile", middleware.RequireAuth(svc), h.UpdateProfile)
	return r, svc
}

func TestIntegration_AuthEmailFlow(t *testing.T) {
	router, _ := newAuthRouter(t)
	email := fmt.Sprintf("handler-%s@example.com", uuid.NewString())

	signupBody := fmt.Sprintf(`{"email":%q,"password":"password123","name":"Handler Test"}`, email)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up/email", bytes.NewBufferString(signupBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("signup %d: %s", w.Code, w.Body.String())
	}

	cookie := w.Header().Get("Set-Cookie")
	if cookie == "" {
		t.Fatal("expected session cookie")
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/auth/get-session", nil)
	req2.Header.Set("Cookie", cookie)
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("session %d: %s", w2.Code, w2.Body.String())
	}

	var sess model.SessionResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &sess); err != nil || sess.User == nil {
		t.Fatalf("bad session: %s", w2.Body.String())
	}

	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/api/auth/sign-out", nil)
	req3.Header.Set("Cookie", cookie)
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("signout %d", w3.Code)
	}
}

func TestAuthHandler_InvalidJSON(t *testing.T) {
	router, _ := newAuthRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up/email", bytes.NewBufferString("{"))
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
}

func TestGoogleSignIn_Redirects(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/sign-in/social?callbackURL=/results", nil)
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)
	cfg := config.AuthConfig{
		FrontendURL:        "http://localhost:3000",
		SessionCookieName:  "realput_session",
		SessionTTL:         time.Hour,
		CookieSecure:       false,
		CookieSameSite:     "lax",
		GoogleClientID:     "test-client",
		GoogleClientSecret: "test-secret",
		GoogleRedirectURL:  "http://localhost:3000/api/auth/callback/google",
		BcryptCost:         10,
	}
	svc := auth.New(store, cfg)
	h := NewAuthHandler(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))
	r := gin.New()
	r.GET("/api/auth/sign-in/social", h.GoogleSignIn)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusFound {
		t.Fatalf("status %d", w.Code)
	}
}

func TestIntegration_UpdateProfile(t *testing.T) {
	router, svc := newAuthRouter(t)
	email := fmt.Sprintf("profile-handler-%s@example.com", uuid.NewString())
	body := fmt.Sprintf(`{"email":%q,"password":"password123","name":"Before"}`, email)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up/email", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	cookie := w.Header().Get("Set-Cookie")
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPatch, "/api/auth/profile", bytes.NewBufferString(`{"name":"After"}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Cookie", cookie)
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("update %d: %s", w2.Code, w2.Body.String())
	}
	_ = svc
}

func TestGoogleCallback_MissingCode(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)
	cfg := config.AuthConfig{
		FrontendURL: "http://localhost:3000", SessionCookieName: "realput_session",
		SessionTTL: time.Hour, CookieSecure: false, CookieSameSite: "lax",
		GoogleClientID: "id", GoogleClientSecret: "secret",
		GoogleRedirectURL: "http://localhost:3000/cb", BcryptCost: 10,
	}
	h := NewAuthHandler(auth.New(store, cfg), slog.New(slog.NewTextHandler(io.Discard, nil)))
	r := gin.New()
	r.GET("/api/auth/callback/google", h.GoogleCallback)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/callback/google?state=ok", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "ok"})
	r.ServeHTTP(w, req)
	if w.Code != http.StatusFound || !strings.Contains(w.Header().Get("Location"), "auth=failed") {
		t.Fatalf("status %d", w.Code)
	}
}

func TestGoogleCallback_InvalidState(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)
	cfg := config.AuthConfig{
		FrontendURL:        "http://localhost:3000",
		SessionCookieName:  "realput_session",
		SessionTTL:         time.Hour,
		CookieSecure:       false,
		CookieSameSite:     "lax",
		GoogleClientID:     "id",
		GoogleClientSecret: "secret",
		GoogleRedirectURL:  "http://localhost:3000/cb",
		BcryptCost:         10,
	}
	h := NewAuthHandler(auth.New(store, cfg), slog.New(slog.NewTextHandler(io.Discard, nil)))
	r := gin.New()
	r.GET("/api/auth/callback/google", h.GoogleCallback)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/callback/google?state=wrong", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusFound || !strings.Contains(w.Header().Get("Location"), "auth=failed") {
		t.Fatalf("status %d loc %q", w.Code, w.Header().Get("Location"))
	}
}

func TestResolveRedirectTarget_ReadsCallbackURL(t *testing.T) {
	got := resolveRedirectTarget("http://localhost:3000", "http://localhost:3000/results")
	if got != "http://localhost:3000/results" {
		t.Fatalf("got %q", got)
	}
}