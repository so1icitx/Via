// Package groq implements guidance generation via Groq Compound (web search).
package groq

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

const apiURL = "https://api.groq.com/openai/v1/chat/completions"

// Client calls Groq models for RealPut BG guidance flows.
type Client struct {
	apiKey       string
	model        string
	resultsModel string
	timeout      time.Duration
	endpoint     string
	http         *http.Client
}

// New creates a Groq API client.
func New(cfg config.GroqConfig) *Client {
	return &Client{
		apiKey:       cfg.APIKey,
		model:        cfg.Model,
		resultsModel: cfg.ResultsModel,
		timeout:      cfg.RequestTimeout,
		endpoint:     apiURL,
		http:         &http.Client{Timeout: cfg.RequestTimeout},
	}
}

func (c *Client) Name() string {
	return "groq"
}

func (c *Client) Validate(ctx context.Context) error {
	_, err := c.chat(ctx, c.model, 0.2, "Reply with valid JSON only: {\"status\":\"ok\"}", "health check")
	if err == nil {
		return nil
	}
	if IsAPIKeyError(err) {
		return fmt.Errorf("GROQ_API_KEY is invalid — check your key at https://console.groq.com/keys")
	}
	if IsBillingError(err) {
		return fmt.Errorf("groq billing: insufficient quota — check https://console.groq.com/settings/billing")
	}
	if IsTransientError(err) || IsRateLimitError(err) {
		return fmt.Errorf("groq temporarily unavailable: %w", err)
	}
	return fmt.Errorf("groq API check failed: %w", err)
}

// GenerateQuestions produces contextual follow-up questions.
func (c *Client) GenerateQuestions(ctx context.Context, req model.QuestionsRequest) (*model.QuestionsResponse, error) {
	if err := guidanceai.ValidateQuestionsRequest(req); err != nil {
		return nil, err
	}

	text, err := c.chat(ctx, c.model, 0.4, guidanceai.QuestionsSystem(req.Lang), guidanceai.QuestionsUser(req.Category, req.UserInput))
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

// GenerateResults uses a two-step pipeline: Compound searches the web, then a fast model emits valid JSON.
func (c *Client) GenerateResults(ctx context.Context, req model.ResultsRequest) (*model.ResultsPayload, error) {
	if err := guidanceai.ValidateResultsRequest(req); err != nil {
		return nil, err
	}

	answersText := strings.Builder{}
	for q, a := range req.Answers {
		fmt.Fprintf(&answersText, "- %s: %s\n", q, a)
	}
	answers := answersText.String()

	research, err := c.chatWithSearch(
		ctx,
		c.resultsModel,
		0.3,
		guidanceai.ResearchSystem(req.Lang),
		guidanceai.ResearchUser(req.Category, req.UserInput, answers),
	)
	if err != nil {
		if IsRateLimitError(err) || IsBillingError(err) {
			research, err = c.chatPlain(ctx, c.model, 0.3,
				guidanceai.ResearchSystem(req.Lang)+"\n\nНяма достъп до търсене в момента — използвай само известни факти за България и посочи „провери на официалния сайт\".",
				guidanceai.ResearchUser(req.Category, req.UserInput, answers),
			)
		}
		if err != nil {
			return nil, err
		}
	}

	system := guidanceai.ResultsSystem(req.Lang)
	user := guidanceai.ResultsFromResearchUser(req.Category, req.UserInput, answers, research)

	text, err := c.chat(ctx, c.model, 0.2, system, user)
	if err != nil {
		return nil, err
	}

	raw, err := guidanceai.ExtractJSON(text)
	if err != nil {
		text, err = c.chat(ctx, c.model, 0.1,
			"Fix the text into ONE valid JSON object only. Escape newlines in strings as \\n. No markdown fences.",
			"Fix into valid JSON:\n"+text,
		)
		if err != nil {
			return nil, err
		}
		raw, err = guidanceai.ExtractJSON(text)
		if err != nil {
			return nil, err
		}
	}

	return guidanceai.ParseResultsPayload(raw)
}

func (c *Client) chatPlain(ctx context.Context, model string, temperature float64, systemPrompt, userPrompt string) (string, error) {
	return c.doChat(ctx, chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: temperature,
	})
}

func (c *Client) chat(ctx context.Context, model string, temperature float64, systemPrompt, userPrompt string) (string, error) {
	return c.doChat(ctx, chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: &responseFormat{Type: "json_object"},
		Temperature:    temperature,
	})
}

func (c *Client) chatWithSearch(ctx context.Context, model string, temperature float64, systemPrompt, userPrompt string) (string, error) {
	return c.doChat(ctx, chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: temperature,
		SearchSettings: &searchSettings{
			Country: "bulgaria",
		},
	})
}

func (c *Client) doChat(ctx context.Context, body chatRequest) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal groq request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("groq request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read groq response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", parseAPIError(resp.StatusCode, raw)
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("parse groq response: %w", err)
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("groq returned empty content")
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

type searchSettings struct {
	Country string `json:"country,omitempty"`
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
	Temperature    float64         `json:"temperature,omitempty"`
	SearchSettings *searchSettings `json:"search_settings,omitempty"`
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
	return fmt.Errorf("groq chat: status %d, message: %s, code: %s, type: %s", status, msg, body.Error.Code, body.Error.Type)
}