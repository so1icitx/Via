package ai

import (
	"strings"

	"github.com/atilatair/realput-bg/backend/internal/gemini"
	"github.com/atilatair/realput-bg/backend/internal/groq"
	"github.com/atilatair/realput-bg/backend/internal/openai"
)

// IsAPIKeyError reports invalid or missing provider API keys.
func IsAPIKeyError(err error) bool {
	return openai.IsAPIKeyError(err) || groq.IsAPIKeyError(err) || gemini.IsAPIKeyError(err)
}

// IsBillingError reports depleted credits or insufficient quota.
func IsBillingError(err error) bool {
	return openai.IsBillingError(err) || groq.IsBillingError(err) || gemini.IsBillingError(err)
}

// IsRateLimitError reports throughput limits.
func IsRateLimitError(err error) bool {
	return openai.IsRateLimitError(err) || groq.IsRateLimitError(err) || gemini.IsRateLimitError(err)
}

// IsTransientError reports temporary outages.
func IsTransientError(err error) bool {
	return openai.IsTransientError(err) || groq.IsTransientError(err) || gemini.IsTransientError(err)
}

// ProviderName extracts the configured provider from an error message when possible.
func ProviderName(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "groq"):
		return "groq"
	case strings.Contains(msg, "gemini"):
		return "gemini"
	case strings.Contains(msg, "openai"):
		return "openai"
	default:
		return ""
	}
}