package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/gemini"
	"github.com/atilatair/realput-bg/backend/internal/groq"
	"github.com/atilatair/realput-bg/backend/internal/openai"
)

// NewProvider creates the configured guidance client.
func NewProvider(ctx context.Context, cfg config.AIConfig) (Provider, error) {
	switch cfg.Provider {
	case "groq":
		if strings.TrimSpace(cfg.Groq.APIKey) == "" {
			return nil, fmt.Errorf("GROQ_API_KEY is required when AI_PROVIDER=groq")
		}
		return groq.New(cfg.Groq), nil
	case "gemini":
		if strings.TrimSpace(cfg.Gemini.APIKey) == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY is required when AI_PROVIDER=gemini")
		}
		return gemini.New(ctx, cfg.Gemini)
	case "openai":
		if strings.TrimSpace(cfg.OpenAI.APIKey) == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required when AI_PROVIDER=openai")
		}
		return openai.New(cfg.OpenAI), nil
	default:
		return nil, fmt.Errorf("unsupported AI_PROVIDER %q (use groq, gemini, or openai)", cfg.Provider)
	}
}

// ValidateProvider checks provider connectivity when possible.
func ValidateProvider(ctx context.Context, provider Provider) error {
	return provider.Validate(ctx)
}