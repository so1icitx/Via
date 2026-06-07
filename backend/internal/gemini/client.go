// Package gemini implements guidance generation via Gemini with Google Search grounding.
package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/guidanceai"
	"github.com/atilatair/realput-bg/backend/internal/model"
	"google.golang.org/genai"
)

// generateFunc is an optional test double for network calls.
type generateFunc func(ctx context.Context, systemPrompt, userPrompt string, withSearch bool) (string, error)

// Client generates structured guidance using Gemini models.
type Client struct {
	api          *genai.Client
	model        string
	timeout      time.Duration
	testGenerate generateFunc
}

// New creates a Gemini API client configured for the Gemini Developer API.
func New(ctx context.Context, cfg config.GeminiConfig) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	return &Client{
		api:     client,
		model:   cfg.Model,
		timeout: cfg.RequestTimeout,
	}, nil
}

func (c *Client) Name() string {
	return "gemini"
}

func (c *Client) Validate(ctx context.Context) error {
	if c.testGenerate != nil {
		_, err := c.testGenerate(ctx, "Reply with valid JSON only: {\"status\":\"ok\"}", "health check", false)
		return c.mapValidateError(err)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.api.Models.GenerateContent(ctx, c.model, genai.Text("Reply with valid JSON only: {\"status\":\"ok\"}"), &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	})
	if err == nil {
		return nil
	}
	_ = resp
	return c.mapValidateError(err)
}

func (c *Client) mapValidateError(err error) error {
	if err == nil {
		return nil
	}
	if IsAPIKeyError(err) {
		return fmt.Errorf("GEMINI_API_KEY is invalid — check your key at https://aistudio.google.com/apikey")
	}
	if IsBillingError(err) {
		return fmt.Errorf("gemini billing: prepayment credits depleted — add credits at https://aistudio.google.com")
	}
	if IsTransientError(err) || IsRateLimitError(err) {
		return fmt.Errorf("gemini temporarily unavailable: %w", err)
	}
	return fmt.Errorf("gemini API check failed: %w", err)
}

// GenerateQuestions produces contextual follow-up questions.
func (c *Client) GenerateQuestions(ctx context.Context, req model.QuestionsRequest) (*model.QuestionsResponse, error) {
	if err := guidanceai.ValidateQuestionsRequest(req); err != nil {
		return nil, err
	}

	text, err := c.generateWithRetry(ctx, guidanceai.QuestionsSystem(req.Lang), guidanceai.QuestionsUser(req.Category, req.UserInput), false)
	if err != nil {
		if IsAPIKeyError(err) || IsBillingError(err) || IsRateLimitError(err) || IsTransientError(err) {
			return nil, err
		}
		return guidanceai.FallbackQuestions(req.Lang), nil
	}

	raw, err := guidanceai.ExtractJSON(text)
	if err != nil {
		return guidanceai.FallbackQuestions(req.Lang), nil
	}

	var out model.QuestionsResponse
	if err := json.Unmarshal(raw, &out); err != nil || len(out.Questions) == 0 {
		return guidanceai.FallbackQuestions(req.Lang), nil
	}

	guidanceai.NormalizeQuestions(&out, req.Lang)
	return &out, nil
}

// GenerateResults uses web search for research, then formats structured JSON without search.
func (c *Client) GenerateResults(ctx context.Context, req model.ResultsRequest) (*model.ResultsPayload, error) {
	if err := guidanceai.ValidateResultsRequest(req); err != nil {
		return nil, err
	}

	answersText := strings.Builder{}
	for q, a := range req.Answers {
		fmt.Fprintf(&answersText, "- %s: %s\n", q, a)
	}
	answers := answersText.String()

	research, err := c.generateWithRetry(ctx, guidanceai.ResearchSystem(req.Lang), guidanceai.ResearchUser(req.Category, req.UserInput, answers), true)
	if err != nil {
		return nil, err
	}

	system := guidanceai.ResultsSystem(req.Lang)
	user := guidanceai.ResultsFromResearchUser(req.Category, req.UserInput, answers, research)

	text, err := c.generateWithRetry(ctx, system, user, false)
	if err != nil {
		return nil, err
	}

	raw, err := guidanceai.ExtractJSON(text)
	if err != nil {
		return nil, err
	}

	return guidanceai.ParseResultsPayload(raw)
}

func (c *Client) generateWithRetry(ctx context.Context, systemPrompt, userPrompt string, withSearch bool) (string, error) {
	const attempts = 3
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			if !sleepContext(ctx, time.Duration(attempt)*3*time.Second) {
				break
			}
		}
		var text string
		var err error
		if c.testGenerate != nil {
			text, err = c.testGenerate(ctx, systemPrompt, userPrompt, withSearch)
		} else {
			text, err = c.generate(ctx, systemPrompt, userPrompt, withSearch)
		}
		if err == nil && text != "" {
			return text, nil
		}
		lastErr = err
		if err == nil {
			lastErr = fmt.Errorf("gemini returned empty content")
		}
		if err != nil && !(IsRateLimitError(err) || IsTransientError(err) || IsEmptyContentError(err)) {
			return "", err
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("gemini returned empty content")
	}
	return "", lastErr
}

func sleepContext(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func (c *Client) generate(ctx context.Context, systemPrompt, userPrompt string, withSearch bool) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	cfg := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: systemPrompt}}},
		MaxOutputTokens:   8192,
	}
	if withSearch {
		cfg.Tools = []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}}
	} else {
		cfg.ResponseMIMEType = "application/json"
	}

	resp, err := c.api.Models.GenerateContent(ctx, c.model, genai.Text(userPrompt), cfg)
	if err != nil {
		return "", fmt.Errorf("gemini generate content: %w", err)
	}

	text := responseText(resp)
	if text == "" {
		return "", fmt.Errorf("gemini returned empty content (%s)", finishReason(resp))
	}
	return text, nil
}

func finishReason(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return "no candidates"
	}
	if resp.Candidates[0].FinishReason == "" {
		return "unknown"
	}
	return string(resp.Candidates[0].FinishReason)
}

func responseText(resp *genai.GenerateContentResponse) string {
	if resp == nil {
		return ""
	}
	var b strings.Builder
	for _, cand := range resp.Candidates {
		if cand == nil || cand.Content == nil {
			continue
		}
		for _, part := range cand.Content.Parts {
			if part == nil || part.Text == "" {
				continue
			}
			b.WriteString(part.Text)
		}
	}
	return strings.TrimSpace(b.String())
}