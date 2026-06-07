package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/atilatair/realput-bg/backend/internal/service/auth"
	"github.com/atilatair/realput-bg/backend/internal/testutil"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHealth(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	Health(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Fatalf("body %v", body)
	}
}

func TestResolveRedirectTarget(t *testing.T) {
	base := "http://localhost:3000"
	tests := []struct {
		in, want string
	}{
		{"", base},
		{"/saved", base + "/saved"},
		{"http://localhost:3000/profile", "http://localhost:3000/profile"},
		{"http://evil.com/phish", base},
		{"javascript:alert(1)", base},
	}
	for _, tc := range tests {
		if got := resolveRedirectTarget(base, tc.in); got != tc.want {
			t.Fatalf("%q => %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestSignUpAndSignInErrors(t *testing.T) {
	if st, _, _ := signUpError(errors.New("email is required")); st != http.StatusBadRequest {
		t.Fatal("expected validation")
	}
	if st, _, code := signInError(auth.ErrInvalidCredentials); st != http.StatusUnauthorized || code != "INVALID_CREDENTIALS" {
		t.Fatalf("unexpected sign-in mapping")
	}
}

func TestGuidanceHandler_Questions_Success(t *testing.T) {
	mock := &testutil.MockProvider{}
	h := NewGuidanceHandler(mock, testLogger())

	body := `{"category":"ученик","userInput":"искам ВУ","lang":"bg"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/questions", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Questions(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if mock.QuestionsCalls != 1 {
		t.Fatal("provider not called")
	}
	var resp model.QuestionsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || len(resp.Questions) == 0 {
		t.Fatalf("bad response: %s", w.Body.String())
	}
}

func TestGuidanceHandler_Questions_ValidationError(t *testing.T) {
	mock := &testutil.MockProvider{QuestionsErr: errors.New("category is required")}
	h := NewGuidanceHandler(mock, testLogger())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/questions", bytes.NewBufferString(`{"category":"ученик","userInput":"x"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Questions(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestGuidanceHandler_Questions_InvalidJSON(t *testing.T) {
	h := NewGuidanceHandler(&testutil.MockProvider{}, testLogger())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/questions", bytes.NewBufferString("{"))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Questions(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
}

func TestGuidanceHandler_Questions_BillingError(t *testing.T) {
	mock := &testutil.MockProvider{QuestionsErr: testutil.Err("gemini", "prepayment credits depleted")}
	h := NewGuidanceHandler(mock, testLogger())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/questions", bytes.NewBufferString(`{"category":"ученик","userInput":"test","lang":"bg"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Questions(c)
	if w.Code != http.StatusPaymentRequired {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestGuidanceHandler_Results_EmptyContent(t *testing.T) {
	mock := &testutil.MockProvider{ResultsErr: errors.New("gemini returned empty content (STOP)")}
	h := NewGuidanceHandler(mock, testLogger())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/results", bytes.NewBufferString(`{"category":"ученик","userInput":"test","lang":"bg","answers":{}}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Results(c)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestGuidanceHandler_Results_Success(t *testing.T) {
	mock := &testutil.MockProvider{}
	h := NewGuidanceHandler(mock, testLogger())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/results", bytes.NewBufferString(`{"category":"ученик","userInput":"test","lang":"bg","answers":{"location":"Пловдив"}}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Results(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestRespondAIError_Matrix(t *testing.T) {
	tests := []struct {
		err    error
		status int
		code   string
	}{
		{testutil.Err("openai", "invalid_api_key"), http.StatusServiceUnavailable, "AI_KEY_INVALID"},
		{testutil.Err("gemini", "quota exceeded billing"), http.StatusPaymentRequired, "GEMINI_BILLING"},
		{testutil.Err("groq", "rate_limit_exceeded 429"), http.StatusTooManyRequests, "AI_RATE_LIMITED"},
		{testutil.Err("gemini", "503 high demand unavailable"), http.StatusServiceUnavailable, "AI_UNAVAILABLE"},
		{testutil.Err("gemini", "empty content STOP"), http.StatusServiceUnavailable, "AI_UNAVAILABLE"},
	}
	for _, tc := range tests {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		if !respondAIError(c, tc.err) {
			t.Fatalf("expected handled error: %v", tc.err)
		}
		if w.Code != tc.status {
			t.Fatalf("%v => status %d want %d", tc.err, w.Code, tc.status)
		}
		var body model.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		if body.Code != tc.code {
			t.Fatalf("%v => code %s want %s", tc.err, body.Code, tc.code)
		}
	}
}