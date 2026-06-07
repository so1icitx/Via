package handler

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/middleware"
	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/atilatair/realput-bg/backend/internal/service/auth"
	"github.com/gin-gonic/gin"
)

// AuthHandler exposes authentication endpoints consumed by the frontend.
type AuthHandler struct {
	auth   *auth.Service
	logger *slog.Logger
}

func NewAuthHandler(authSvc *auth.Service, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{auth: authSvc, logger: logger}
}

func (h *AuthHandler) SignUpEmail(c *gin.Context) {
	var req model.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request", Code: "INVALID_JSON"})
		return
	}

	user, token, err := h.auth.SignUpEmail(c.Request.Context(), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		status, message, code := signUpError(err)
		c.JSON(status, model.ErrorResponse{Error: message, Code: code})
		return
	}

	setSessionCookie(c, h.auth.CookieConfig(), token)
	c.JSON(http.StatusOK, model.SessionResponse{User: user})
}

func (h *AuthHandler) SignInEmail(c *gin.Context) {
	var req model.SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request", Code: "INVALID_JSON"})
		return
	}

	user, token, err := h.auth.SignInEmail(c.Request.Context(), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		status, message, code := signInError(err)
		c.JSON(status, model.ErrorResponse{Error: message, Code: code})
		return
	}

	setSessionCookie(c, h.auth.CookieConfig(), token)
	c.JSON(http.StatusOK, model.SessionResponse{User: user})
}

func (h *AuthHandler) GetSession(c *gin.Context) {
	token := readSessionToken(c, h.auth.CookieConfig().SessionCookieName)
	sess, err := h.auth.GetSession(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "session lookup failed", Code: "SESSION_ERROR"})
		return
	}
	c.JSON(http.StatusOK, sess)
}

func (h *AuthHandler) SignOut(c *gin.Context) {
	token := readSessionToken(c, h.auth.CookieConfig().SessionCookieName)
	_ = h.auth.SignOut(c.Request.Context(), token)
	clearSessionCookie(c, h.auth.CookieConfig())
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized", Code: "UNAUTHORIZED"})
		return
	}

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request", Code: "INVALID_JSON"})
		return
	}

	user, err := h.auth.UpdateProfile(c.Request.Context(), userID, req.Name)
	if err != nil {
		status, message, code := updateProfileError(err)
		c.JSON(status, model.ErrorResponse{Error: message, Code: code})
		return
	}
	c.JSON(http.StatusOK, model.SessionResponse{User: user})
}

func (h *AuthHandler) GoogleSignIn(c *gin.Context) {
	state, err := randomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "oauth state error", Code: "OAUTH_ERROR"})
		return
	}

	cfg := h.auth.CookieConfig()
	setOAuthStateCookie(c, cfg, state)

	callback := resolveRedirectTarget(cfg.FrontendURL, c.Query("callbackURL"))
	setOAuthCallbackCookie(c, cfg, callback)

	c.Redirect(http.StatusFound, h.auth.GoogleAuthURL(state))
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	cfg := h.auth.CookieConfig()
	stateCookie, _ := c.Cookie("oauth_state")
	if stateCookie == "" || stateCookie != c.Query("state") {
		c.Redirect(http.StatusFound, cfg.FrontendURL+"?auth=failed")
		return
	}
	clearOAuthStateCookie(c, cfg)

	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusFound, cfg.FrontendURL+"?auth=failed")
		return
	}

	_, token, err := h.auth.CompleteGoogleOAuth(c.Request.Context(), code, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		h.logger.Error("google oauth failed", "error", err)
		c.Redirect(http.StatusFound, cfg.FrontendURL+"?auth=failed")
		return
	}

	setSessionCookie(c, cfg, token)
	callback, _ := c.Cookie("oauth_callback")
	callback, _ = url.QueryUnescape(callback)
	clearOAuthCallbackCookie(c, cfg)
	c.Redirect(http.StatusFound, resolveRedirectTarget(cfg.FrontendURL, callback))
}

func setSessionCookie(c *gin.Context, cfg config.AuthConfig, token string) {
	c.SetSameSite(parseSameSite(cfg.CookieSameSite))
	c.SetCookie(cfg.SessionCookieName, token, int(cfg.SessionTTL.Seconds()), "/", "", cfg.CookieSecure, true)
}

func clearSessionCookie(c *gin.Context, cfg config.AuthConfig) {
	c.SetSameSite(parseSameSite(cfg.CookieSameSite))
	c.SetCookie(cfg.SessionCookieName, "", -1, "/", "", cfg.CookieSecure, true)
}

func setOAuthStateCookie(c *gin.Context, cfg config.AuthConfig, state string) {
	c.SetSameSite(parseSameSite(cfg.CookieSameSite))
	c.SetCookie("oauth_state", state, 600, "/", "", cfg.CookieSecure, true)
}

func clearOAuthStateCookie(c *gin.Context, cfg config.AuthConfig) {
	c.SetCookie("oauth_state", "", -1, "/", "", cfg.CookieSecure, true)
}

func setOAuthCallbackCookie(c *gin.Context, cfg config.AuthConfig, callback string) {
	c.SetSameSite(parseSameSite(cfg.CookieSameSite))
	c.SetCookie("oauth_callback", callback, 600, "/", "", cfg.CookieSecure, true)
}

func clearOAuthCallbackCookie(c *gin.Context, cfg config.AuthConfig) {
	c.SetCookie("oauth_callback", "", -1, "/", "", cfg.CookieSecure, true)
}

func readSessionToken(c *gin.Context, cookieName string) string {
	if token, err := c.Cookie(cookieName); err == nil {
		return token
	}
	return c.GetHeader("X-Session-Token")
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func parseSameSite(value string) http.SameSite {
	switch strings.ToLower(value) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}