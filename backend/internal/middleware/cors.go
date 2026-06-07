package middleware

import (
	"net/http"
	"strings"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/gin-gonic/gin"
)

// CORS applies configurable cross-origin rules with credential support.
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		allowed[origin] = struct{}{}
	}
	_, allowAll := allowed["*"]

	methods := strings.Join(cfg.AllowedMethods, ", ")
	headers := strings.Join(cfg.AllowedHeaders, ", ")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if allowAll {
				c.Header("Access-Control-Allow-Origin", "*")
			} else if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Vary", "Origin")
			}
			c.Header("Access-Control-Allow-Methods", methods)
			c.Header("Access-Control-Allow-Headers", headers)
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}