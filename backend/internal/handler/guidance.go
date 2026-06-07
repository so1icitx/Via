package handler

import (
	"log/slog"
	"net/http"

	"github.com/atilatair/realput-bg/backend/internal/ai"
	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/gin-gonic/gin"
)

// GuidanceHandler serves the AI question and results endpoints.
type GuidanceHandler struct {
	provider ai.Provider
	logger   *slog.Logger
}

func NewGuidanceHandler(provider ai.Provider, logger *slog.Logger) *GuidanceHandler {
	return &GuidanceHandler{provider: provider, logger: logger}
}

// Questions generates follow-up questions for the guidance flow.
func (h *GuidanceHandler) Questions(c *gin.Context) {
	var req model.QuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "invalid request body",
			Code:  "INVALID_JSON",
		})
		return
	}

	resp, err := h.provider.GenerateQuestions(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("generate questions failed", "error", err)
		if respondAIError(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "invalid request",
			Code:  "VALIDATION_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Results generates personalized guidance.
func (h *GuidanceHandler) Results(c *gin.Context) {
	var req model.ResultsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "invalid request body",
			Code:  "INVALID_JSON",
		})
		return
	}

	resp, err := h.provider.GenerateResults(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("generate results failed", "error", err)
		if respondAIError(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: "Възникна грешка при генериране на резултатите.",
			Code:  "GENERATION_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}