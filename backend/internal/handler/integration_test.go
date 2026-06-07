package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/middleware"
	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/atilatair/realput-bg/backend/internal/testutil"
	"github.com/gin-gonic/gin"
)

func newTestRouter(t *testing.T, provider *testutil.MockProvider) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CORS(config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Origin"},
	}))
	r.Use(middleware.RateLimit(config.RateLimitConfig{RequestsPerSecond: 1000, Burst: 1000}))

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	guidance := NewGuidanceHandler(provider, logger)
	r.GET("/health", Health)
	api := r.Group("/api")
	{
		api.POST("/questions", guidance.Questions)
		api.POST("/results", guidance.Results)
	}
	return r
}

func TestIntegration_GuidanceFlow(t *testing.T) {
	provider := &testutil.MockProvider{
		QuestionsResponse: &model.QuestionsResponse{
			Questions: []model.Question{
				{ID: "education_level", Question: "Клас?", Type: "choice", Options: []string{"12 клас"}},
				{ID: "location", Question: "Град?", Type: "choice", Options: []string{"Пловдив"}},
			},
		},
	}
	router := newTestRouter(t, provider)

	// Step 1: questions
	qBody := `{"category":"ученик","userInput":"мрежова сигурност","lang":"bg"}`
	req := httptest.NewRequest(http.MethodPost, "/api/questions", bytes.NewBufferString(qBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("questions status %d: %s", w.Code, w.Body.String())
	}

	// Step 2: results
	rBody := `{"category":"ученик","userInput":"мрежова сигурност","lang":"bg","answers":{"education_level":"12 клас","location":"Пловдив"}}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/results", bytes.NewBufferString(rBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("results status %d: %s", w2.Code, w2.Body.String())
	}

	var payload model.ResultsPayload
	if err := json.Unmarshal(w2.Body.Bytes(), &payload); err != nil || len(payload.Items) == 0 {
		t.Fatalf("bad results payload: %s", w2.Body.String())
	}
	if provider.QuestionsCalls != 1 || provider.ResultsCalls != 1 {
		t.Fatalf("calls questions=%d results=%d", provider.QuestionsCalls, provider.ResultsCalls)
	}
}

func TestIntegration_CORS_Preflight(t *testing.T) {
	router := newTestRouter(t, &testutil.MockProvider{})
	req := httptest.NewRequest(http.MethodOptions, "/api/questions", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatal("missing CORS header")
	}
}