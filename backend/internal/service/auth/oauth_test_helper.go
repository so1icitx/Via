package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/repository"
	"github.com/google/uuid"
)

type oauthMockProfile struct {
	ID    string
	Email string
	Name  string
}

func newOAuthMockServer(t *testing.T, profile *oauthMockProfile, tokenStatus, profileStatus int) (*httptest.Server, *oauthMockProfile) {
	t.Helper()
	if profile == nil {
		profile = &oauthMockProfile{}
	}
	if tokenStatus == 0 {
		tokenStatus = http.StatusOK
	}
	if profileStatus == 0 {
		profileStatus = http.StatusOK
	}
	if profile.ID == "" {
		profile.ID = "google-id-" + uuid.NewString()
	}
	if profile.Email == "" {
		profile.Email = fmt.Sprintf("oauth-%s@example.com", uuid.NewString())
	}
	if profile.Name == "" {
		profile.Name = "OAuth User"
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(tokenStatus)
			if tokenStatus == http.StatusOK {
				_, _ = w.Write([]byte(`{"access_token":"test-access-token","token_type":"bearer"}`))
			} else {
				_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
			}
		case "/userinfo":
			w.WriteHeader(profileStatus)
			if profileStatus == http.StatusOK {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id": profile.ID, "email": profile.Email,
					"name": profile.Name, "verified_email": true,
				})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	return srv, profile
}

func oauthTestService(t *testing.T, store *repository.Store, srv *httptest.Server) *Service {
	t.Helper()
	cfg := config.AuthConfig{
		SessionCookieName:  "realput_session",
		SessionTTL:         0, // set by caller if needed
		GoogleClientID:     "client-id",
		GoogleClientSecret: "client-secret",
		GoogleRedirectURL:  "http://localhost:3000/cb",
		BcryptCost:         10,
	}
	svc := New(store, cfg)
	svc.GoogleOAuthTestEndpoints(srv.URL+"/token", srv.URL+"/userinfo")
	return svc
}