package handler

import (
	"net/http"
	"strings"

	"github.com/atilatair/realput-bg/backend/internal/ai"
	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/gin-gonic/gin"
)

func respondAIError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if ai.IsAPIKeyError(err) {
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{
			Error: "AI service is not configured. Check the API key for your AI_PROVIDER in backend/.env.",
			Code:  "AI_KEY_INVALID",
		})
		return true
	}
	if ai.IsBillingError(err) {
		msg := "AI credits are depleted. Add billing for your configured provider."
		code := "AI_BILLING"
		switch ai.ProviderName(err) {
		case "gemini":
			msg = "Gemini credits are depleted. Add credits at https://aistudio.google.com"
			code = "GEMINI_BILLING"
		case "groq":
			msg = "Groq quota exhausted. Check https://console.groq.com/settings/billing"
			code = "GROQ_BILLING"
		case "openai":
			msg = "OpenAI credits are depleted. Add billing at https://platform.openai.com/settings/organization/billing"
			code = "OPENAI_BILLING"
		}
		c.JSON(http.StatusPaymentRequired, model.ErrorResponse{
			Error: msg,
			Code:  code,
		})
		return true
	}
	if ai.IsRateLimitError(err) {
		c.JSON(http.StatusTooManyRequests, model.ErrorResponse{
			Error: "AI rate limit reached. Wait a minute and try again.",
			Code:  "AI_RATE_LIMITED",
		})
		return true
	}
	if ai.IsTransientError(err) {
		msg := "AI is temporarily busy. Please try again in a moment."
		if strings.Contains(strings.ToLower(err.Error()), "empty content") {
			msg = "AI не върна пълен отговор (Gemini беше претоварен). Изчакай малко и опитай отново — резултатите отнемат 1–2 минути."
		}
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{
			Error: msg,
			Code:  "AI_UNAVAILABLE",
		})
		return true
	}
	return false
}