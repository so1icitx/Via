// Package testutil provides shared test doubles for integration tests.
package testutil

import (
	"context"
	"fmt"

	"github.com/atilatair/realput-bg/backend/internal/guidanceai"
	"github.com/atilatair/realput-bg/backend/internal/model"
)

// MockProvider implements ai.Provider for handler tests.
type MockProvider struct {
	NameValue            string
	ValidateErr          error
	QuestionsErr         error
	ResultsErr           error
	QuestionsResponse    *model.QuestionsResponse
	ResultsResponse      *model.ResultsPayload
	QuestionsCalls       int
	ResultsCalls         int
	LastQuestionsRequest model.QuestionsRequest
	LastResultsRequest   model.ResultsRequest
}

func (m *MockProvider) Name() string {
	if m.NameValue != "" {
		return m.NameValue
	}
	return "mock"
}

func (m *MockProvider) Validate(ctx context.Context) error {
	return m.ValidateErr
}

func (m *MockProvider) GenerateQuestions(ctx context.Context, req model.QuestionsRequest) (*model.QuestionsResponse, error) {
	m.QuestionsCalls++
	m.LastQuestionsRequest = req
	if m.QuestionsErr != nil {
		return nil, m.QuestionsErr
	}
	if m.QuestionsResponse != nil {
		return m.QuestionsResponse, nil
	}
	return guidanceai.FallbackQuestions(req.Lang), nil
}

func (m *MockProvider) GenerateResults(ctx context.Context, req model.ResultsRequest) (*model.ResultsPayload, error) {
	m.ResultsCalls++
	m.LastResultsRequest = req
	if m.ResultsErr != nil {
		return nil, m.ResultsErr
	}
	if m.ResultsResponse != nil {
		return m.ResultsResponse, nil
	}
	return &model.ResultsPayload{
		Intro: "test intro",
		Items: []model.ResultItem{
			{ID: "path-1", Title: "Test University — CS", Summary: "fits"},
		},
		Opportunities: []model.Opportunity{
			{Title: "Test Hackathon", Type: "Хакатон"},
		},
	}, nil
}

// Err returns a formatted error tagged with a provider name.
func Err(provider, msg string) error {
	return fmt.Errorf("%s: %s", provider, msg)
}