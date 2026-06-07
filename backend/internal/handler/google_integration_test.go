package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/repository"
	"github.com/atilatair/realput-bg/backend/internal/service/auth"
	"github.com/atilatair/realput-bg/backend/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestGoogleCallback_SuccessRedirect(t *testing.T) {
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)

	oauthSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok","token_type":"bearer"}`))
		case "/userinfo":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "gid-" + uuid.NewString(),
				"email": fmt.Sprintf("cb-%s@example.com", uuid.NewString()),
				"name": "CB User", "verified_email": true,
			})
		}
	}))
	defer oauthSrv.Close()

	cfg := config.AuthConfig{
		FrontendURL: "http://localhost:3000", SessionCookieName: "realput_session",
		SessionTTL: time.Hour, CookieSecure: false, CookieSameSite: "lax",
		GoogleClientID: "id", GoogleClientSecret: "secret",
		GoogleRedirectURL: "http://localhost:3000/cb", BcryptCost: 10,
	}
	svc := auth.New(store, cfg)
	svc.GoogleOAuthTestEndpoints(oauthSrv.URL+"/token", oauthSrv.URL+"/userinfo")

	h := NewAuthHandler(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))
	r := gin.New()
	r.GET("/api/auth/callback/google", h.GoogleCallback)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/callback/google?state=ok&code=abc", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "ok"})
	req.AddCookie(&http.Cookie{Name: "oauth_callback", Value: "http%3A%2F%2Flocalhost%3A3000%2Fresults"})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("status %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc == "" || loc == "http://localhost:3000?auth=failed" {
		t.Fatalf("location %q", loc)
	}
}