package openai

import (
	"errors"
	"testing"
)

func TestOpenAIErrorHelpers(t *testing.T) {
	cases := map[string]func(error) bool{
		"invalid_api_key":          IsAPIKeyError,
		"insufficient_quota":       IsBillingError,
		"rate_limit_exceeded":      IsRateLimitError,
		"503 temporarily unavailable": IsTransientError,
	}
	for msg, fn := range cases {
		if !fn(errors.New("openai chat: " + msg)) {
			t.Fatalf("expected true for %q", msg)
		}
	}
	if IsRateLimitError(errors.New("insufficient_quota")) {
		t.Fatal("billing should not be rate limit")
	}
}