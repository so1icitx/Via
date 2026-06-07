package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/gin-gonic/gin"
)

func TestParseSameSite(t *testing.T) {
	if parseSameSite("strict") != http.SameSiteStrictMode {
		t.Fatal("strict")
	}
	if parseSameSite("none") != http.SameSiteNoneMode {
		t.Fatal("none")
	}
	if parseSameSite("lax") != http.SameSiteLaxMode {
		t.Fatal("lax")
	}
}

func TestSessionCookies(t *testing.T) {
	cfg := config.AuthConfig{
		SessionCookieName: "realput_session",
		SessionTTL:        time.Hour,
		CookieSecure:      false,
		CookieSameSite:    "lax",
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	setSessionCookie(c, cfg, "tok123")
	if w.Header().Get("Set-Cookie") == "" {
		t.Fatal("expected set-cookie")
	}
	clearSessionCookie(c, cfg)
}

func TestRandomState(t *testing.T) {
	a, err := randomState()
	if err != nil || len(a) < 10 {
		t.Fatalf("state %q err %v", a, err)
	}
	b, _ := randomState()
	if a == b {
		t.Fatal("expected unique state")
	}
}

func TestOAuthCookies(t *testing.T) {
	cfg := config.AuthConfig{CookieSecure: false, CookieSameSite: "lax"}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	setOAuthStateCookie(c, cfg, "state1")
	setOAuthCallbackCookie(c, cfg, "http://localhost:3000/")
	clearOAuthStateCookie(c, cfg)
	clearOAuthCallbackCookie(c, cfg)
}