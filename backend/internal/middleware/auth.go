package middleware

import (
	"net/http"

	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/atilatair/realput-bg/backend/internal/service/auth"
	"github.com/gin-gonic/gin"
)

const userIDKey = "userID"

// RequireAuth ensures a valid session exists before protected handlers run.
func RequireAuth(authSvc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := readSessionToken(c, authSvc.CookieConfig().SessionCookieName)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
				Error: "unauthorized",
				Code:  "UNAUTHORIZED",
			})
			return
		}

		sess, err := authSvc.GetSession(c.Request.Context(), token)
		if err != nil || sess.User == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
				Error: "unauthorized",
				Code:  "UNAUTHORIZED",
			})
			return
		}

		c.Set(userIDKey, sess.User.ID)
		c.Next()
	}
}

// UserID returns the authenticated user id set by RequireAuth.
func UserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(userIDKey)
	if !ok {
		return "", false
	}
	id, ok := v.(string)
	return id, ok && id != ""
}

func readSessionToken(c *gin.Context, cookieName string) string {
	if token, err := c.Cookie(cookieName); err == nil && token != "" {
		return token
	}
	return c.GetHeader("X-Session-Token")
}