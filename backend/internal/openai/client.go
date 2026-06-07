// Package openai implements guidance generation via the OpenAI Chat Completions API.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/guidanceai"
	"github.com/atilatair/realput-bg/backend/internal/model"
)

const apiURL = "https://api.openai.com/v1/chat/completions"

// Client calls OpenAI models for RealPut BG guidance flows.
type Client struct {
	apiKey          string
	model           string
	resultsModel    string
	webSearch       bool
	timeout         time.Duration
	endpoint        string
	responsesEndpoint string
	http            *http.Client
}

// New creates an OpenAI API client.
func New(cfg config.OpenAIConfig) *Client {
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 180 * time.Second
	}
	resultsModel := cfg.ResultsModel
	if strings.TrimSpace(resultsModel) == "" {
		resultsModel = cfg.Model
	}
	return &Client{
		apiKey:       cfg.APIKey,
		model:        cfg.Model,
		resultsModel: resultsModel,
		webSearch:    cfg.WebSearch,
		timeout:      timeout,
		endpoint:          apiURL,
		responsesEndpoint: responsesURL,
		http:              &http.Client{Timeout: timeout},
	}
}

func (c *Client) Name() string {
	return "openai"
}

func (c *Client) Validate(ctx context.Context) error {
	_, err := c.chat(ctx, "Reply with valid JSON only: {\"status\":\"ok\"}", "health check")
	if err == nil {
		return nil
	}
	if IsAPIKeyError(err) {
		return fmt.Errorf("OPENAI_API_KEY is invalid — check your key at https://platform.openai.com/api-keys")
	}
	if IsBillingError(err) {
		return fmt.Errorf("openai billing: insufficient quota — add credits at https://platform.openai.com/settings/organization/billing")
	}
	if IsTransientError(err) || IsRateLimitError(err) {
		return fmt.Errorf("openai temporarily unavailable: %w", err)
	}
	return fmt.Errorf("openai API check failed: %w", err)
}

// GenerateQuestions produces contextual follow-up questions.
func (c *Client) GenerateQuestions(ctx context.Context, req model.QuestionsRequest) (*model.QuestionsResponse, error) {
	if err := guidanceai.ValidateQuestionsRequest(req); err != nil {
		return nil, err
	}

	text, err := c.chat(ctx, guidanceai.QuestionsSystem(req.Lang), guidanceai.QuestionsUser(req.Category, req.UserInput))
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

// GenerateResults produces personalized guidance.
func (c *Client) GenerateResults(ctx context.Context, req model.ResultsRequest) (*model.ResultsPayload, error) {
	if err := guidanceai.ValidateResultsRequest(req); err != nil {
		return nil, err
	}

	answersText := strings.Builder{}
	for q, a := range req.Answers {
		fmt.Fprintf(&answersText, "- %s: %s\n", q, a)
	}

	system := guidanceai.ResultsSystem(req.Lang)
	user := guidanceai.ResultsUser(req.Category, req.UserInput, answersText.String())

	var text string
	var err error
	if c.webSearch {
		text, err = c.chatWithWebSearch(ctx, c.resultsModel, 0.2, system, user)
		if err != nil {
			text, err = c.chatWithModel(ctx, c.resultsModel, 0.2, system, user)
		}
	} else {
		text, err = c.chatWithModel(ctx, c.resultsModel, 0.2, system, user)
	}
	if err != nil {
		return nil, err
	}

	raw, err := guidanceai.ExtractJSON(text)
	if err != nil {
		return nil, err
	}

	return guidanceai.ParseResultsPayload(raw)
}

func (c *Client) chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return c.chatWithModel(ctx, c.model, 0.4, systemPrompt, userPrompt)
}

func (c *Client) chatWithModel(ctx context.Context, model string, temperature float64, systemPrompt, userPrompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	body := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: &responseFormat{Type: "json_object"},
		Temperature:    temperature,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read openai response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", parseAPIError(resp.StatusCode, raw)
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("parse openai response: %w", err)
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("openai returned empty content")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
	Temperature    float64         `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *apiErrorBody `json:"error"`
}

type apiErrorBody struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func parseAPIError(status int, raw []byte) error {
	var body struct {
		Error apiErrorBody `json:"error"`
	}
	_ = json.Unmarshal(raw, &body)
	msg := strings.TrimSpace(body.Error.Message)
	if msg == "" {
		msg = strings.TrimSpace(string(raw))
	}
	return fmt.Errorf("openai chat: status %d, message: %s, code: %s, type: %s", status, msg, body.Error.Code, body.Error.Type)
}