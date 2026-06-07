package config

import (
	"testing"
	"time"
)

func setBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://realput:realput@localhost:5432/realput?sslmode=disable")
	t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
}

func TestLoad_GroqProvider(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("AI_PROVIDER", "groq")
	t.Setenv("GROQ_API_KEY", "gsk_test_key")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AI.Provider != "groq" {
		t.Fatalf("provider %q", cfg.AI.Provider)
	}
	if cfg.Server.Port != "8080" {
		t.Fatalf("port %q", cfg.Server.Port)
	}
}

func TestLoad_GeminiProviderRequiresKey(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("AI_PROVIDER", "gemini")
	t.Setenv("GEMINI_API_KEY", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing gemini key error")
	}
}

func TestLoad_UnsupportedProvider(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("AI_PROVIDER", "anthropic")
	t.Setenv("GROQ_API_KEY", "gsk_test")

	if _, err := Load(); err == nil {
		t.Fatal("expected unsupported provider error")
	}
}

func TestLoad_InvalidBcryptCost(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("AI_PROVIDER", "groq")
	t.Setenv("GROQ_API_KEY", "gsk_test")
	t.Setenv("BCRYPT_COST", "99")

	if _, err := Load(); err == nil {
		t.Fatal("expected bcrypt validation error")
	}
}

func TestDetectAIProvider_PrefersGroq(t *testing.T) {
	if got := detectAIProvider("groq", "gem", "oa"); got != "groq" {
		t.Fatalf("got %q", got)
	}
}

func TestDetectAIProvider_Fallbacks(t *testing.T) {
	if got := detectAIProvider("", "gem", "oa"); got != "gemini" {
		t.Fatalf("got %q", got)
	}
	if got := detectAIProvider("", "", "oa"); got != "openai" {
		t.Fatalf("got %q", got)
	}
	if got := detectAIProvider("", "", ""); got != "groq" {
		t.Fatalf("got %q", got)
	}
}

func TestLoad_AutoDetectGemini(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("AI_PROVIDER", "")
	t.Setenv("GROQ_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "gem_test")
	t.Setenv("OPENAI_API_KEY", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AI.Provider != "gemini" {
		t.Fatalf("provider %q", cfg.AI.Provider)
	}
}

func TestLoad_OpenAIProvider(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("AI_PROVIDER", "openai")
	t.Setenv("OPENAI_API_KEY", "sk_test")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AI.Provider != "openai" {
		t.Fatalf("provider %q", cfg.AI.Provider)
	}
}

func TestDefaultSameSite(t *testing.T) {
	if defaultSameSite("http://a", "http://b") != "lax" {
		t.Fatal("expected lax for http")
	}
	if defaultSameSite("https://api.example.com", "https://app.example.com") != "none" {
		t.Fatal("expected none for https")
	}
}

func TestLoad_InvalidRateLimit(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("AI_PROVIDER", "groq")
	t.Setenv("GROQ_API_KEY", "gsk_test")
	t.Setenv("RATE_LIMIT_RPS", "0")

	if _, err := Load(); err == nil {
		t.Fatal("expected rate limit error")
	}
}

func TestEnvStringSlice(t *testing.T) {
	t.Setenv("TEST_SLICE", " a , b , , c ")
	got := envStringSlice("TEST_SLICE", []string{"default"})
	if len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("got %#v", got)
	}
}

func TestEnvHelpers(t *testing.T) {
	t.Setenv("TEST_INT", "42")
	if envInt("TEST_INT", 0) != 42 {
		t.Fatal("envInt")
	}
	t.Setenv("TEST_BOOL", "true")
	if !envBool("TEST_BOOL", false) {
		t.Fatal("envBool")
	}
	t.Setenv("TEST_DURATION", "5s")
	if envDuration("TEST_DURATION", 0) != 5*time.Second {
		t.Fatal("envDuration")
	}
	if envString("MISSING_KEY_XYZ", "fallback") != "fallback" {
		t.Fatal("envString")
	}
}