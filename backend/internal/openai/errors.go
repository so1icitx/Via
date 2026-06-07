package openai

import "strings"

// IsAPIKeyError reports invalid or missing OpenAI API keys.
func IsAPIKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid_api_key") ||
		strings.Contains(msg, "incorrect api key") ||
		strings.Contains(msg, "invalid openai")
}

// IsBillingError reports depleted credits or insufficient quota.
func IsBillingError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "insufficient_quota") ||
		strings.Contains(msg, "billing") ||
		strings.Contains(msg, "exceeded your current quota") ||
		strings.Contains(msg, "insufficient quota")
}

// IsRateLimitError reports throughput limits.
func IsRateLimitError(err error) bool {
	if err == nil || IsBillingError(err) {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "rate_limit") ||
		strings.Contains(msg, "429")
}

// IsTransientError reports temporary outages.
func IsTransientError(err error) bool {
	if err == nil || IsBillingError(err) || IsRateLimitError(err) {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "503") ||
		strings.Contains(msg, "500") ||
		strings.Contains(msg, "overloaded") ||
		strings.Contains(msg, "temporarily unavailable")
}