package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const responsesURL = "https://api.openai.com/v1/responses"

type webSearchTool struct {
	Type         string              `json:"type"`
	UserLocation *webSearchLocation  `json:"user_location,omitempty"`
	SearchContextSize string         `json:"search_context_size,omitempty"`
}

type webSearchLocation struct {
	Type    string `json:"type"`
	Country string `json:"country,omitempty"`
	City    string `json:"city,omitempty"`
}

type responsesInputMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responsesRequest struct {
	Model       string                  `json:"model"`
	Input       []responsesInputMessage `json:"input"`
	Tools       []webSearchTool         `json:"tools,omitempty"`
	Temperature float64                 `json:"temperature,omitempty"`
}

type responsesResponse struct {
	OutputText string                `json:"output_text"`
	Output     []responsesOutputItem `json:"output"`
	Error      *apiErrorBody         `json:"error"`
}

type responsesOutputItem struct {
	Type    string                 `json:"type"`
	Content []responsesContentPart `json:"content"`
}

type responsesContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (c *Client) chatWithWebSearch(ctx context.Context, model string, temperature float64, systemPrompt, userPrompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	body := responsesRequest{
		Model: model,
		Input: []responsesInputMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Tools: []webSearchTool{
			{
				Type: "web_search",
				UserLocation: &webSearchLocation{
					Type:    "approximate",
					Country: "BG",
				},
				SearchContextSize: "high",
			},
		},
		Temperature: temperature,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal openai responses request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.responsesEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai responses request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read openai responses body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", parseAPIError(resp.StatusCode, raw)
	}

	var parsed responsesResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("parse openai responses: %w", err)
	}

	text := extractResponsesText(parsed)
	if text == "" {
		return "", fmt.Errorf("openai responses returned empty content")
	}
	return text, nil
}

func extractResponsesText(parsed responsesResponse) string {
	if t := strings.TrimSpace(parsed.OutputText); t != "" {
		return t
	}

	var b strings.Builder
	for _, item := range parsed.Output {
		if item.Type != "message" {
			continue
		}
		for _, part := range item.Content {
			if part.Type != "output_text" {
				continue
			}
			partText := strings.TrimSpace(part.Text)
			if partText == "" {
				continue
			}
			if b.Len() > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(partText)
		}
	}
	return strings.TrimSpace(b.String())
}