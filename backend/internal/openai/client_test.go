package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/model"
)

func mockOpenAIServer(t *testing.T, content string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if status >= 400 {
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{"message": "invalid_api_key", "type": "invalid_request_error"},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []struct {
				Message chatMessage `json:"message"`
			}{{Message: chatMessage{Role: "assistant", Content: content}}},
		})
	}))
}

func testClient(srv *httptest.Server) *Client {
	c := New(config.OpenAIConfig{
		APIKey:         "test-key",
		Model:          "gpt-4o-mini",
		ResultsModel:   "gpt-4o-mini",
		RequestTimeout: 5 * time.Second,
		WebSearch:      false,
	})
	c.endpoint = srv.URL
	c.http = srv.Client()
	return c
}

func TestClient_Name(t *testing.T) {
	c := New(config.OpenAIConfig{APIKey: "sk-test", Model: "gpt-4o-mini"})
	if c.Name() != "openai" {
		t.Fatalf("name %q", c.Name())
	}
}

func TestGenerateQuestions_Success(t *testing.T) {
	srv := mockOpenAIServer(t, `{"questions":[{"id":"education_level","question":"Age?","type":"choice","options":["12 клас"]}]}`, http.StatusOK)
	defer srv.Close()

	c := testClient(srv)
	resp, err := c.GenerateQuestions(context.Background(), model.QuestionsRequest{
		Category: "ученик", UserInput: "искам ВУ", Lang: "bg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Questions) == 0 || resp.Questions[0].ID != "education_level" {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestGenerateQuestions_FallbackOnBadJSON(t *testing.T) {
	srv := mockOpenAIServer(t, "not json", http.StatusOK)
	defer srv.Close()

	c := testClient(srv)
	resp, err := c.GenerateQuestions(context.Background(), model.QuestionsRequest{
		Category: "ученик", UserInput: "test", Lang: "bg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Questions[0].ID != "education_level" {
		t.Fatal("expected fallback with education_level first")
	}
}

func TestGenerateQuestions_APIKeyError(t *testing.T) {
	srv := mockOpenAIServer(t, "", http.StatusUnauthorized)
	defer srv.Close()

	c := testClient(srv)
	_, err := c.GenerateQuestions(context.Background(), model.QuestionsRequest{
		Category: "ученик", UserInput: "test", Lang: "bg",
	})
	if err == nil || !IsAPIKeyError(err) {
		t.Fatalf("expected api key error, got %v", err)
	}
}

func TestGenerateResults_Success(t *testing.T) {
	payload := `{"intro":"hi","items":[{"id":"p1","title":"TU Sofia — CS","summary":"ok","whyFits":"- yes","opportunities":"details","honestAssessment":"- hard","nextSteps":"- apply"}],"opportunities":[{"title":"Event","type":"Събитие","description":"d","howToApply":"a","timing":"soon"}]}`
	srv := mockOpenAIServer(t, payload, http.StatusOK)
	defer srv.Close()

	c := testClient(srv)
	resp, err := c.GenerateResults(context.Background(), model.ResultsRequest{
		Category: "ученик", UserInput: "test", Lang: "bg",
		Answers: map[string]string{"location": "София"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("items %d", len(resp.Items))
	}
}

func TestGenerateResults_WebSearchPath(t *testing.T) {
	payload := `{"intro":"ok","items":[{"id":"p1","title":"TU","summary":"s","whyFits":"-","opportunities":"-","honestAssessment":"-","nextSteps":"-"}],"opportunities":[]}`
	searchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"output_text": payload})
	}))
	defer searchSrv.Close()

	c := New(config.OpenAIConfig{
		APIKey: "test-key", Model: "gpt-4o-mini", ResultsModel: "gpt-4o-mini",
		RequestTimeout: 5 * time.Second, WebSearch: true,
	})
	c.responsesEndpoint = searchSrv.URL
	c.http = searchSrv.Client()

	resp, err := c.GenerateResults(context.Background(), model.ResultsRequest{
		Category: "ученик", UserInput: "test", Lang: "bg", Answers: map[string]string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("items=%d", len(resp.Items))
	}
}

func TestValidate_Success(t *testing.T) {
	srv := mockOpenAIServer(t, `{"status":"ok"}`, http.StatusOK)
	defer srv.Close()
	c := testClient(srv)
	if err := c.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestParseAPIError(t *testing.T) {
	err := parseAPIError(429, []byte(`{"error":{"message":"rate_limit","code":"rate_limit_exceeded","type":"rate_limit"}}`))
	if err == nil || !IsRateLimitError(err) {
		t.Fatalf("got %v", err)
	}
}