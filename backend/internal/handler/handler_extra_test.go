package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atilatair/realput-bg/backend/internal/service/auth"
	"github.com/atilatair/realput-bg/backend/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestIntegration_SignInEmail(t *testing.T) {
	router, _ := newAuthRouter(t)
	email := fmt.Sprintf("signin-%s@example.com", uuid.NewString())

	signup := fmt.Sprintf(`{"email":%q,"password":"password123","name":"Sign In"}`, email)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up/email", bytes.NewBufferString(signup))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("signup %d: %s", w.Code, w.Body.String())
	}

	signin := fmt.Sprintf(`{"email":%q,"password":"password123"}`, email)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/sign-in/email", bytes.NewBufferString(signin))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("signin %d: %s", w2.Code, w2.Body.String())
	}
}

func TestSignInEmail_InvalidCredentials(t *testing.T) {
	router, _ := newAuthRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-in/email", bytes.NewBufferString(`{"email":"nobody@example.com","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", w.Code)
	}
}

func TestSignUpEmail_Duplicate(t *testing.T) {
	router, _ := newAuthRouter(t)
	email := fmt.Sprintf("dup-%s@example.com", uuid.NewString())
	body := fmt.Sprintf(`{"email":%q,"password":"password123","name":"Dup"}`, email)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up/email", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up/email", bytes.NewBufferString(body))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Fatalf("status %d body %s", w2.Code, w2.Body.String())
	}
}

func TestUpdateProfile_ValidationError(t *testing.T) {
	router, _ := newAuthRouter(t)
	email := fmt.Sprintf("profile-val-%s@example.com", uuid.NewString())
	signup := fmt.Sprintf(`{"email":%q,"password":"password123","name":"X"}`, email)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up/email", bytes.NewBufferString(signup))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPatch, "/api/auth/profile", bytes.NewBufferString(`{"name":"   "}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Cookie", w.Header().Get("Set-Cookie"))
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w2.Code, w2.Body.String())
	}
}

func TestErrorMappings(t *testing.T) {
	st, msg, code := signUpError(auth.ErrEmailExists)
	if st != http.StatusConflict || code != "EMAIL_EXISTS" {
		t.Fatalf("signup duplicate: %d %s %s", st, msg, code)
	}
	st, _, code = signInError(auth.ErrInvalidCredentials)
	if st != http.StatusUnauthorized || code != "INVALID_CREDENTIALS" {
		t.Fatalf("signin invalid: %d %s", st, code)
	}
	st, _, code = updateProfileError(fmt.Errorf("name is required"))
	if st != http.StatusBadRequest || code != "VALIDATION_ERROR" {
		t.Fatalf("profile validation: %d %s", st, code)
	}
	st, _, code = updateProfileError(fmt.Errorf("db down"))
	if st != http.StatusBadRequest || code != "UPDATE_FAILED" {
		t.Fatalf("profile generic: %d %s", st, code)
	}
}

func TestGuidanceHandler_Results_InvalidJSON(t *testing.T) {
	h := NewGuidanceHandler(&testutil.MockProvider{}, testLogger())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/results", bytes.NewBufferString("{"))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Results(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
}

func TestGuidanceHandler_Results_GenericError(t *testing.T) {
	mock := &testutil.MockProvider{ResultsErr: fmt.Errorf("unexpected failure")}
	h := NewGuidanceHandler(mock, testLogger())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/results", bytes.NewBufferString(`{"category":"ученик","userInput":"x","lang":"bg","answers":{}}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Results(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}