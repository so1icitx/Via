package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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

func newSavedRouter(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	pool := testutil.OpenPostgres(t)
	store := repository.New(pool)
	cfg := config.AuthConfig{
		SessionCookieName:  "realput_session",
		SessionTTL:         time.Hour,
		CookieSecure:       false,
		CookieSameSite:     "lax",
		GoogleClientID:     "id",
		GoogleClientSecret: "secret",
		GoogleRedirectURL:  "http://localhost:3000/cb",
		BcryptCost:         10,
	}
	svc := auth.New(store, cfg)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	email := fmt.Sprintf("saved-handler-%s@example.com", uuid.NewString())
	user, token, err := svc.SignUpEmail(context.Background(), model.SignUpRequest{
		Email: email, Password: "password123", Name: "Saver",
	}, "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	_ = user

	saved := NewSavedHandler(store, logger)
	r := gin.New()
	protected := r.Group("/api")
	protected.Use(middleware.RequireAuth(svc))
	{
		protected.GET("/saved-results", saved.List)
		protected.POST("/saved-results", saved.Create)
		protected.DELETE("/saved-results/:id", saved.Delete)
	}
	return r, token
}

func TestIntegration_SavedResultsHandlers(t *testing.T) {
	router, token := newSavedRouter(t)

	createBody := `{"query":"uni search","data":{"intro":"x","items":[{"id":"p1","title":"TU","summary":"s"}],"opportunities":[]}}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/saved-results", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-Token", token)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create %d: %s", w.Code, w.Body.String())
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/saved-results", nil)
	req2.Header.Set("X-Session-Token", token)
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("list %d: %s", w2.Code, w2.Body.String())
	}

	var rows []model.SavedResultRow
	if err := json.Unmarshal(w2.Body.Bytes(), &rows); err != nil || len(rows) == 0 {
		t.Fatalf("rows: %s", w2.Body.String())
	}

	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/saved-results/%d", rows[0].ID), nil)
	req3.Header.Set("X-Session-Token", token)
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusNoContent {
		t.Fatalf("delete %d", w3.Code)
	}
}

func TestSavedHandler_CreateValidation(t *testing.T) {
	router, token := newSavedRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/saved-results", bytes.NewBufferString(`{"query":"","data":{"intro":"x","items":[],"opportunities":[]}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-Token", token)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
}

func TestSavedHandler_Unauthorized(t *testing.T) {
	h := NewSavedHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/saved-results", nil)
	h.List(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", w.Code)
	}
}