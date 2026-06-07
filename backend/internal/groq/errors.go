package groq

import "strings"

// IsAPIKeyError reports invalid or missing Groq API keys.
func IsAPIKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid api key") ||
		strings.Contains(msg, "invalid_api_key") ||
		strings.Contains(msg, "groq_api_key")
}

// IsBillingError reports depleted credits or insufficient quota.
func IsBillingError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "insufficient") ||
		strings.Contains(msg, "billing") ||
		strings.Contains(msg, "quota") ||
		strings.Contains(msg, "credits")
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