package guidanceai

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/atilatair/realput-bg/backend/internal/model"
)

// ValidateQuestionsRequest checks question generation input.
func ValidateQuestionsRequest(req model.QuestionsRequest) error {
	if strings.TrimSpace(req.Category) == "" {
		return fmt.Errorf("category is required")
	}
	if strings.TrimSpace(req.UserInput) == "" {
		return fmt.Errorf("userInput is required")
	}
	if utf8.RuneCountInString(req.UserInput) > 4000 {
		return fmt.Errorf("userInput is too long")
	}
	return nil
}

// ValidateResultsRequest checks results generation input.
func ValidateResultsRequest(req model.ResultsRequest) error {
	if strings.TrimSpace(req.Category) == "" {
		return fmt.Errorf("category is required")
	}
	if strings.TrimSpace(req.UserInput) == "" {
		return fmt.Errorf("userInput is required")
	}
	if utf8.RuneCountInString(req.UserInput) > 4000 {
		return fmt.Errorf("userInput is too long")
	}
	return nil
}