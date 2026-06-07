package ai

import (
	"context"

	"github.com/atilatair/realput-bg/backend/internal/model"
)

// Provider generates guidance questions and results.
type Provider interface {
	Name() string
	Validate(ctx context.Context) error
	GenerateQuestions(ctx context.Context, req model.QuestionsRequest) (*model.QuestionsResponse, error)
	GenerateResults(ctx context.Context, req model.ResultsRequest) (*model.ResultsPayload, error)
}