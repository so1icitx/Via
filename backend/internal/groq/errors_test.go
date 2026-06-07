package groq

import (
	"errors"
	"testing"
)

func TestGroqErrorHelpers(t *testing.T) {
	if !IsAPIKeyError(errors.New("groq chat: invalid api key")) {
		t.Fatal("api key")
	}
	if !IsBillingError(errors.New("groq chat: insufficient quota")) {
		t.Fatal("billing")
	}
	if !IsRateLimitError(errors.New("groq chat: rate_limit_exceeded 429")) {
		t.Fatal("rate")
	}
	if !IsTransientError(errors.New("groq chat: 503 overloaded")) {
		t.Fatal("transient")
	}
}