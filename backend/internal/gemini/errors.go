package gemini

import "strings"

// IsAPIKeyError reports invalid or missing Gemini API keys.
func IsAPIKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "api key not valid") ||
		strings.Contains(msg, "api_key_invalid") ||
		strings.Contains(msg, "invalid api key") ||
		strings.Contains(msg, "gemini_api_key")
}

// IsBillingError reports depleted credits or insufficient quota.
func IsBillingError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "quota") ||
		strings.Contains(msg, "billing") ||
		strings.Contains(msg, "credits depleted") ||
		strings.Contains(msg, "prepayment") ||
		strings.Contains(msg, "exceeded your current quota")
}

// IsRateLimitError reports throughput limits.
func IsRateLimitError(err error) bool {
	if err == nil || IsBillingError(err) {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "resource_exhausted")
}

// IsTransientError reports temporary outages.
func IsTransientError(err error) bool {
	if err == nil || IsBillingError(err) || IsRateLimitError(err) {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "503") ||
		strings.Contains(msg, "500") ||
		strings.Contains(msg, "unavailable") ||
		strings.Contains(msg, "high demand") ||
		IsEmptyContentError(err)
}

// IsEmptyContentError reports when the model finished without returning text.
func IsEmptyContentError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "empty content")
}

