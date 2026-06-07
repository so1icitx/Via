package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimit applies per-IP token bucket limits.
func RateLimit(cfg config.RateLimitConfig) gin.HandlerFunc {
	store := &visitorStore{
		limiters: make(map[string]*rate.Limiter),
		lastSeen: make(map[string]time.Time),
		rps:      rate.Limit(cfg.RequestsPerSecond),
		burst:    cfg.Burst,
	}

	return func(c *gin.Context) {
		ip := clientIP(c)
		if !store.get(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, model.ErrorResponse{
				Error: "rate limit exceeded",
				Code:  "RATE_LIMITED",
			})
			return
		}
		c.Next()
	}
}

type visitorStore struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	lastSeen map[string]time.Time
	rps      rate.Limit
	burst    int
}

func (s *visitorStore) get(key string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.lastSeen[key] = now
	if len(s.limiters) > 2048 {
		cutoff := now.Add(-time.Hour)
		for ip, seen := range s.lastSeen {
			if seen.Before(cutoff) {
				delete(s.limiters, ip)
				delete(s.lastSeen, ip)
			}
		}
	}

	lim, ok := s.limiters[key]
	if !ok {
		lim = rate.NewLimiter(s.rps, s.burst)
		s.limiters[key] = lim
	}
	return lim
}

func clientIP(c *gin.Context) string {
	if c.GetBool("trust_proxy") {
		if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
			parts := strings.SplitN(forwarded, ",", 2)
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
	}
	return c.ClientIP()
}