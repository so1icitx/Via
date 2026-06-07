// Package config loads runtime settings from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config aggregates all service configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	AI       AIConfig
	Auth     AuthConfig
	CORS     CORSConfig
	Rate     RateLimitConfig
}

type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	MaxRequestBytes int64
	TrustProxy      bool
}

type DatabaseConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MigrateOnStart  bool
}

// AIConfig selects and configures the guidance AI provider.
type AIConfig struct {
	Provider       string
	RequestTimeout time.Duration
	OpenAI         OpenAIConfig
	Groq           GroqConfig
	Gemini         GeminiConfig
}

// OpenAIConfig configures the OpenAI guidance client.
type OpenAIConfig struct {
	APIKey         string
	Model          string
	ResultsModel   string
	WebSearch      bool
	RequestTimeout time.Duration
}

// GroqConfig configures Groq Compound models with built-in web search.
type GroqConfig struct {
	APIKey       string
	Model        string
	ResultsModel string
	RequestTimeout time.Duration
}

// GeminiConfig configures Gemini with Google Search grounding.
type GeminiConfig struct {
	APIKey       string
	Model        string
	RequestTimeout time.Duration
}

type AuthConfig struct {
	BaseURL            string
	FrontendURL        string
	SessionCookieName  string
	SessionTTL         time.Duration
	SessionUpdateAge   time.Duration
	CookieSecure       bool
	CookieSameSite     string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	BcryptCost         int
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         time.Duration
}

type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
}

// Load reads and validates configuration from the environment.
func Load() (*Config, error) {
	_ = godotenv.Load()

	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	googleID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	googleSecret := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
	if googleID == "" || googleSecret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET are required")
	}

	frontendURL := envString("FRONTEND_URL", "http://localhost:3000")
	baseURL := envString("AUTH_BASE_URL", envString("API_BASE_URL", "http://localhost:8080"))
	aiTimeout := envDuration("AI_REQUEST_TIMEOUT", envDuration("OPENAI_REQUEST_TIMEOUT", 180*time.Second))

	openaiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	groqKey := strings.TrimSpace(os.Getenv("GROQ_API_KEY"))
	geminiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	provider := strings.ToLower(envString("AI_PROVIDER", detectAIProvider(groqKey, geminiKey, openaiKey)))

	cfg := &Config{
		Server: ServerConfig{
			Host:            envString("SERVER_HOST", "0.0.0.0"),
			Port:            envString("SERVER_PORT", "8080"),
			ReadTimeout:     envDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    envDuration("SERVER_WRITE_TIMEOUT", 180*time.Second),
			IdleTimeout:     envDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			MaxRequestBytes: envInt64("SERVER_MAX_REQUEST_BYTES", 1<<20),
			TrustProxy:      envBool("TRUST_PROXY", false),
		},
		Database: DatabaseConfig{
			URL:             dbURL,
			MaxConns:        int32(envInt("DB_MAX_CONNS", 10)),
			MinConns:        int32(envInt("DB_MIN_CONNS", 2)),
			MaxConnLifetime: envDuration("DB_MAX_CONN_LIFETIME", 30*time.Minute),
			MigrateOnStart:  envBool("DB_MIGRATE_ON_START", true),
		},
		AI: AIConfig{
			Provider:       provider,
			RequestTimeout: aiTimeout,
			OpenAI: OpenAIConfig{
				APIKey:         openaiKey,
				Model:          envString("OPENAI_MODEL", "gpt-4o-mini"),
				ResultsModel:   envString("OPENAI_RESULTS_MODEL", "gpt-4o"),
				WebSearch:      envBool("OPENAI_WEB_SEARCH", true),
				RequestTimeout: aiTimeout,
			},
			Groq: GroqConfig{
				APIKey:         groqKey,
				Model:          envString("GROQ_MODEL", "llama-3.1-8b-instant"),
				ResultsModel:   envString("GROQ_RESULTS_MODEL", "groq/compound-mini"),
				RequestTimeout: aiTimeout,
			},
			Gemini: GeminiConfig{
				APIKey:         geminiKey,
				Model:          envString("GEMINI_MODEL", "gemini-2.5-flash"),
				RequestTimeout: aiTimeout,
			},
		},
		Auth: AuthConfig{
			BaseURL:            strings.TrimRight(baseURL, "/"),
			FrontendURL:        strings.TrimRight(frontendURL, "/"),
			SessionCookieName:  envString("SESSION_COOKIE_NAME", "realput_session"),
			SessionTTL:         envDuration("SESSION_TTL", 7*24*time.Hour),
			SessionUpdateAge:   envDuration("SESSION_UPDATE_AGE", 24*time.Hour),
			CookieSecure:       envBool("COOKIE_SECURE", strings.HasPrefix(baseURL, "https://")),
			CookieSameSite:     envString("COOKIE_SAME_SITE", defaultSameSite(baseURL, frontendURL)),
			GoogleClientID:     googleID,
			GoogleClientSecret: googleSecret,
			GoogleRedirectURL:  envString("GOOGLE_REDIRECT_URL", strings.TrimRight(frontendURL, "/")+"/api/auth/callback/google"),
			BcryptCost:         envInt("BCRYPT_COST", 12),
		},
		CORS: CORSConfig{
			AllowedOrigins: envStringSlice("CORS_ALLOWED_ORIGINS", []string{frontendURL}),
			AllowedMethods: envStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"}),
			AllowedHeaders: envStringSlice("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization", "Cookie"}),
			MaxAge:         envDuration("CORS_MAX_AGE", 12*time.Hour),
		},
		Rate: RateLimitConfig{
			RequestsPerSecond: envFloat("RATE_LIMIT_RPS", 30),
			Burst:             envInt("RATE_LIMIT_BURST", 60),
		},
	}

	if cfg.Rate.RequestsPerSecond <= 0 || cfg.Rate.Burst < 1 {
		return nil, fmt.Errorf("invalid rate limit configuration")
	}
	if cfg.Auth.BcryptCost < 10 || cfg.Auth.BcryptCost > 14 {
		return nil, fmt.Errorf("BCRYPT_COST must be between 10 and 14")
	}

	switch cfg.AI.Provider {
	case "groq":
		if strings.TrimSpace(cfg.AI.Groq.APIKey) == "" {
			return nil, fmt.Errorf("GROQ_API_KEY is required when AI_PROVIDER=groq (free key at https://console.groq.com/keys)")
		}
	case "gemini":
		if strings.TrimSpace(cfg.AI.Gemini.APIKey) == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY is required when AI_PROVIDER=gemini")
		}
	case "openai":
		if strings.TrimSpace(cfg.AI.OpenAI.APIKey) == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required when AI_PROVIDER=openai")
		}
	default:
		return nil, fmt.Errorf("unsupported AI_PROVIDER %q (use groq, gemini, or openai)", cfg.AI.Provider)
	}

	return cfg, nil
}

func detectAIProvider(groqKey, geminiKey, openaiKey string) string {
	if groqKey != "" {
		return "groq"
	}
	if geminiKey != "" {
		return "gemini"
	}
	if openaiKey != "" {
		return "openai"
	}
	return "groq"
}

func defaultSameSite(apiURL, frontendURL string) string {
	if strings.HasPrefix(apiURL, "https://") || strings.HasPrefix(frontendURL, "https://") {
		return "none"
	}
	return "lax"
}

func envString(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envInt64(key string, fallback int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return n
}

func envFloat(key string, fallback float64) float64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func envBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func envStringSlice(key string, fallback []string) []string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}