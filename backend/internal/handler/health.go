package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health reports service readiness for orchestrators.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "realput-bg",
	})
}