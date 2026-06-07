package ai

import (
	"context"
	"errors"
	"testing"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/testutil"
)

func TestErrorClassifiers(t *testing.T) {
	tests := []struct {
		err      error
		key      bool
		billing  bool
		rate     bool
		trans    bool
		provider string
	}{
		{errors.New("openai chat: invalid_api_key"), true, false, false, false, "openai"},
		{errors.New("gemini billing: quota exceeded"), false, true, false, false, "gemini"},
		{errors.New("groq chat: rate_limit_exceeded 429"), false, false, true, false, "groq"},
		{errors.New("gemini 503 unavailable high demand"), false, false, false, true, "gemini"},
		{errors.New("gemini returned empty content"), false, false, false, true, "gemini"},
	}
	for _, tc := range tests {
		if IsAPIKeyError(tc.err) != tc.key {
			t.Fatalf("key %v for %v", tc.err, tc.key)
		}
		if IsBillingError(tc.err) != tc.billing {
			t.Fatalf("billing %v for %v", tc.err, tc.billing)
		}
		if IsRateLimitError(tc.err) != tc.rate {
			t.Fatalf("rate %v for %v", tc.err, tc.rate)
		}
		if IsTransientError(tc.err) != tc.trans {
			t.Fatalf("transient %v for %v", tc.err, tc.trans)
		}
		if ProviderName(tc.err) != tc.provider {
			t.Fatalf("provider %q for %v", ProviderName(tc.err), tc.err)
		}
	}
}

func TestNewProvider_Unsupported(t *testing.T) {
	_, err := NewProvider(context.Background(), config.AIConfig{Provider: "fake"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewProvider_GroqMissingKey(t *testing.T) {
	_, err := NewProvider(context.Background(), config.AIConfig{Provider: "groq"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateProvider_Mock(t *testing.T) {
	p := &testutil.MockProvider{}
	if err := ValidateProvider(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	p.ValidateErr = errors.New("gemini billing: quota")
	if err := ValidateProvider(context.Background(), p); err == nil {
		t.Fatal("expected validate error")
	}
}

func TestNewProvider_GroqSuccess(t *testing.T) {
	p, err := NewProvider(context.Background(), config.AIConfig{
		Provider: "groq",
		Groq: config.GroqConfig{
			APIKey:         "gsk_test",
			Model:          "llama-3.1-8b-instant",
			ResultsModel:   "groq/compound-mini",
			RequestTimeout: 0,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "groq" {
		t.Fatalf("name %q", p.Name())
	}
}

func TestNewProvider_GeminiMissingKey(t *testing.T) {
	_, err := NewProvider(context.Background(), config.AIConfig{Provider: "gemini"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewProvider_OpenAIMissingKey(t *testing.T) {
	_, err := NewProvider(context.Background(), config.AIConfig{Provider: "openai"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewProvider_GeminiSuccess(t *testing.T) {
	p, err := NewProvider(context.Background(), config.AIConfig{
		Provider: "gemini",
		Gemini:   config.GeminiConfig{APIKey: "test-key", Model: "gemini-2.5-flash"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "gemini" {
		t.Fatalf("name %q", p.Name())
	}
}

func TestNewProvider_OpenAISuccess(t *testing.T) {
	p, err := NewProvider(context.Background(), config.AIConfig{
		Provider: "openai",
		OpenAI: config.OpenAIConfig{
			APIKey: "sk-test", Model: "gpt-4o-mini", ResultsModel: "gpt-4o",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "openai" {
		t.Fatalf("name %q", p.Name())
	}
}

func TestProviderName_Unknown(t *testing.T) {
	if ProviderName(errors.New("something else")) != "" {
		t.Fatal("expected empty provider")
	}
}