package gemini

import (
	"errors"
	"testing"
)

func TestGeminiErrorHelpers(t *testing.T) {
	if !IsAPIKeyError(errors.New("gemini generate content: API_KEY_INVALID")) {
		t.Fatal("api key")
	}
	if !IsBillingError(errors.New("gemini billing: quota exceeded prepayment")) {
		t.Fatal("billing")
	}
	if !IsRateLimitError(errors.New("gemini generate content: 429 resource_exhausted")) {
		t.Fatal("rate")
	}
	if !IsTransientError(errors.New("gemini 503 high demand unavailable")) {
		t.Fatal("transient")
	}
	if !IsEmptyContentError(errors.New("gemini returned empty content (STOP)")) {
		t.Fatal("empty content")
	}
	if IsRateLimitError(errors.New("gemini billing: quota exceeded")) {
		t.Fatal("billing should not count as rate limit")
	}
}