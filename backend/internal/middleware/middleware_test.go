package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/atilatair/realput-bg/backend/internal/service/auth"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCORS_AllowedOrigin(t *testing.T) {
	r := gin.New()
	r.Use(CORS(config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	}))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatal("missing allow-origin")
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("missing credentials header")
	}
}

func TestCORS_OptionsAborts(t *testing.T) {
	r := gin.New()
	r.Use(CORS(config.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	}))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusTeapot) })

	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status %d", w.Code)
	}
}

func TestRateLimit_BlocksBurst(t *testing.T) {
	r := gin.New()
	r.Use(RateLimit(config.RateLimitConfig{RequestsPerSecond: 1, Burst: 1}))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	var lastCode int
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
		lastCode = w.Code
	}
	if lastCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", lastCode)
	}
}

func TestClientIP_TrustProxy(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")
	c.Set("trust_proxy", true)
	if got := clientIP(c); got != "203.0.113.1" {
		t.Fatalf("got %q", got)
	}
}

func TestRequireAuth_NoCookie(t *testing.T) {
	svc := auth.New(nil, config.AuthConfig{SessionCookieName: "realput_session"})
	r := gin.New()
	r.Use(RequireAuth(svc))
	r.GET("/secure", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/secure", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", w.Code)
	}
}

func TestVisitorStore_Eviction(t *testing.T) {
	store := &visitorStore{
		limiters: make(map[string]*rate.Limiter),
		lastSeen: make(map[string]time.Time),
		rps:      10,
		burst:    10,
	}
	// Fill beyond cleanup threshold with old timestamps.
	for i := 0; i < 2050; i++ {
		key := fmt.Sprintf("ip-%d", i)
		store.lastSeen[key] = time.Now().Add(-2 * time.Hour)
		store.limiters[key] = rate.NewLimiter(10, 10)
	}
	_ = store.get("fresh-ip")
	if len(store.limiters) >= 2050 {
		t.Fatal("expected stale limiter eviction")
	}
}

func TestUserID_Missing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if _, ok := UserID(c); ok {
		t.Fatal("expected false")
	}
	c.Set(userIDKey, "abc")
	if id, ok := UserID(c); !ok || id != "abc" {
		t.Fatalf("got %q %v", id, ok)
	}
}

func TestReadSessionToken_HeaderFallback(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("X-Session-Token", "tok123")
	if got := readSessionToken(c, "realput_session"); got != "tok123" {
		t.Fatalf("got %q", got)
	}
}

// Ensure model import used for rate limit JSON shape compile-time check.
var _ = model.ErrorResponse{}