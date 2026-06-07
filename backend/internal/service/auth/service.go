// Package auth implements session-based authentication and Google OAuth.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/atilatair/realput-bg/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrUnauthorized       = errors.New("unauthorized")
)

const defaultGoogleProfileURL = "https://www.googleapis.com/oauth2/v2/userinfo"

// Service coordinates user authentication flows.
type Service struct {
	store             *repository.Store
	cfg               config.AuthConfig
	google            *oauth2.Config
	googleProfileURL  string
}

func New(store *repository.Store, cfg config.AuthConfig) *Service {
	return &Service{
		store:            store,
		cfg:              cfg,
		googleProfileURL: defaultGoogleProfileURL,
		google: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

// GoogleOAuthTestEndpoints overrides OAuth endpoints for integration tests.
func (s *Service) GoogleOAuthTestEndpoints(tokenURL, profileURL string) {
	authURL := strings.TrimSuffix(tokenURL, "/token") + "/auth"
	s.google = &oauth2.Config{
		ClientID:     s.cfg.GoogleClientID,
		ClientSecret: s.cfg.GoogleClientSecret,
		RedirectURL:  s.cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}
	s.googleProfileURL = profileURL
}

func (s *Service) SignUpEmail(ctx context.Context, req model.SignUpRequest, ip, userAgent string) (*model.AuthUser, string, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)
	password := req.Password

	if email == "" || password == "" {
		return nil, "", fmt.Errorf("email and password are required")
	}
	if len(password) < 8 {
		return nil, "", fmt.Errorf("password must be at least 8 characters")
	}
	if name == "" {
		name = strings.Split(email, "@")[0]
	}

	existing, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", ErrEmailExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.BcryptCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	user, err := s.store.CreateUser(ctx, name, email, false)
	if err != nil {
		return nil, "", err
	}
	if err := s.store.CreateCredentialAccount(ctx, user.ID, string(hash)); err != nil {
		return nil, "", err
	}

	token, err := s.issueSession(ctx, user.ID, ip, userAgent)
	if err != nil {
		return nil, "", err
	}

	return toAuthUser(user), token, nil
}

func (s *Service) SignInEmail(ctx context.Context, req model.SignInRequest, ip, userAgent string) (*model.AuthUser, string, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		return nil, "", ErrInvalidCredentials
	}

	user, err := s.store.FindUserByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, "", ErrInvalidCredentials
	}

	hash, err := s.store.GetCredentialPasswordHash(ctx, user.ID)
	if err != nil || hash == "" {
		return nil, "", ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.issueSession(ctx, user.ID, ip, userAgent)
	if err != nil {
		return nil, "", err
	}
	return toAuthUser(user), token, nil
}

func (s *Service) GoogleAuthURL(state string) string {
	return s.google.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (s *Service) CompleteGoogleOAuth(ctx context.Context, code, ip, userAgent string) (*model.AuthUser, string, error) {
	token, err := s.google.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("exchange google code: %w", err)
	}

	client := s.google.Client(ctx, token)
	resp, err := client.Get(s.googleProfileURL)
	if err != nil {
		return nil, "", fmt.Errorf("fetch google profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("google profile status: %d", resp.StatusCode)
	}

	var profile struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		VerifiedEmail bool   `json:"verified_email"`
	}
	if err := decodeJSON(resp.Body, &profile); err != nil {
		return nil, "", err
	}
	if profile.Email == "" || profile.ID == "" {
		return nil, "", fmt.Errorf("incomplete google profile")
	}

	user, err := s.store.FindGoogleAccount(ctx, profile.ID)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		byEmail, err := s.store.FindUserByEmail(ctx, strings.ToLower(profile.Email))
		if err != nil {
			return nil, "", err
		}
		if byEmail != nil {
			user = byEmail
		} else {
			created, err := s.store.CreateUser(ctx, profile.Name, strings.ToLower(profile.Email), true)
			if err != nil {
				return nil, "", err
			}
			user = created
		}

		if err := s.store.CreateGoogleAccount(ctx, user.ID, profile.ID); err != nil {
			return nil, "", err
		}
	}

	sessionToken, err := s.issueSession(ctx, user.ID, ip, userAgent)
	if err != nil {
		return nil, "", err
	}
	return toAuthUser(user), sessionToken, nil
}

func (s *Service) GetSession(ctx context.Context, token string) (*model.SessionResponse, error) {
	sess, err := s.store.FindSessionByToken(ctx, token)
	if err != nil || sess == nil {
		return &model.SessionResponse{}, nil
	}
	if time.Now().UTC().After(sess.ExpiresAt) {
		_ = s.store.DeleteSessionByToken(ctx, token)
		return &model.SessionResponse{}, nil
	}

	user, err := s.store.FindUserByID(ctx, sess.UserID)
	if err != nil || user == nil {
		return &model.SessionResponse{}, nil
	}

	_ = s.store.TouchSession(ctx, sess.ID)

	return &model.SessionResponse{
		User: toAuthUser(user),
		Session: &struct {
			ID        string `json:"id"`
			ExpiresAt string `json:"expiresAt"`
		}{
			ID:        sess.ID,
			ExpiresAt: sess.ExpiresAt.UTC().Format(time.RFC3339),
		},
	}, nil
}

func (s *Service) SignOut(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.store.DeleteSessionByToken(ctx, token)
}

func (s *Service) UpdateProfile(ctx context.Context, userID, name string) (*model.AuthUser, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if err := s.store.UpdateUserName(ctx, userID, name); err != nil {
		return nil, err
	}
	user, err := s.store.FindUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrUnauthorized
	}
	return toAuthUser(user), nil
}

func (s *Service) CookieConfig() config.AuthConfig {
	return s.cfg
}

func (s *Service) issueSession(ctx context.Context, userID, ip, userAgent string) (string, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}
	_, err = s.store.CreateSession(ctx, userID, token, ip, userAgent, time.Now().UTC().Add(s.cfg.SessionTTL))
	if err != nil {
		return "", err
	}
	return token, nil
}

func toAuthUser(u *repository.UserRecord) *model.AuthUser {
	return &model.AuthUser{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Image:         u.Image,
	}
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}