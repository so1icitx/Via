// Command api boots the RealPut BG (РеалПът БГ) backend API.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/ai"
	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/db"
	"github.com/atilatair/realput-bg/backend/internal/handler"
	"github.com/atilatair/realput-bg/backend/internal/middleware"

	"github.com/atilatair/realput-bg/backend/internal/repository"
	"github.com/atilatair/realput-bg/backend/internal/service/auth"
	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration error", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.Database)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	aiProvider, err := ai.NewProvider(ctx, cfg.AI)
	if err != nil {
		logger.Error("AI provider init failed", "provider", cfg.AI.Provider, "error", err)
		os.Exit(1)
	}
	if err := ai.ValidateProvider(ctx, aiProvider); err != nil {
		if ai.IsBillingError(err) || ai.IsTransientError(err) || ai.IsRateLimitError(err) {
			logger.Warn("AI API check skipped", "provider", cfg.AI.Provider, "error", err)
		} else {
			logger.Error("AI API check failed", "provider", cfg.AI.Provider, "error", err)
			os.Exit(1)
		}
	} else {
		logger.Info("AI API ready", "provider", cfg.AI.Provider, "name", aiProvider.Name())
	}

	store := repository.New(pool.Pool)
	authSvc := auth.New(store, cfg.Auth)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(maxBodySize(cfg.Server.MaxRequestBytes))
	router.Use(requestLogger(logger))
	router.Use(middleware.CORS(cfg.CORS))
	router.Use(func(c *gin.Context) {
		c.Set("trust_proxy", cfg.Server.TrustProxy)
		c.Next()
	})

	guidance := handler.NewGuidanceHandler(aiProvider, logger)
	authHandler := handler.NewAuthHandler(authSvc, logger)
	savedHandler := handler.NewSavedHandler(store, logger)

	router.GET("/health", handler.Health)

	api := router.Group("/api")
	aiRoutes := api.Group("")
	aiRoutes.Use(middleware.RateLimit(cfg.Rate))
	{
		aiRoutes.POST("/questions", guidance.Questions)
		aiRoutes.POST("/results", guidance.Results)
	}

	authRoutes := api.Group("/auth")
	{
		authRoutes.POST("/sign-up/email", authHandler.SignUpEmail)
		authRoutes.POST("/sign-in/email", authHandler.SignInEmail)
		authRoutes.GET("/get-session", authHandler.GetSession)
		authRoutes.POST("/sign-out", authHandler.SignOut)
		authRoutes.GET("/sign-in/social", authHandler.GoogleSignIn)
		authRoutes.GET("/callback/google", authHandler.GoogleCallback)
	}

	protected := api.Group("")
	protected.Use(middleware.RequireAuth(authSvc))
	{
		protected.PATCH("/auth/profile", authHandler.UpdateProfile)
		protected.GET("/saved-results", savedHandler.List)
		protected.POST("/saved-results", savedHandler.Create)
		protected.DELETE("/saved-results/:id", savedHandler.Delete)
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:           router,
		ReadTimeout:       cfg.Server.ReadTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		logger.Info("server started", "addr", srv.Addr, "provider", cfg.AI.Provider)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
	}
}

func maxBodySize(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > limit {
			c.AbortWithStatus(http.StatusRequestEntityTooLarge)
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}