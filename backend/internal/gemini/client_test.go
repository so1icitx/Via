package gemini

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/atilatair/realput-bg/backend/internal/config"
	"github.com/atilatair/realput-bg/backend/internal/model"
	"google.golang.org/genai"
)

func TestNew_AndName(t *testing.T) {
	c, err := New(context.Background(), config.GeminiConfig{
		APIKey: "test-key", Model: "gemini-2.5-flash", RequestTimeout: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.Name() != "gemini" {
		t.Fatalf("name %q", c.Name())
	}
}

func testGeminiClient(fn generateFunc) *Client {
	return &Client{
		model:        "gemini-2.5-flash",
		timeout:      5 * time.Second,
		testGenerate: fn,
	}
}

func TestGenerateQuestions_WithStub(t *testing.T) {
	c := testGeminiClient(func(ctx context.Context, system, user string, withSearch bool) (string, error) {
		return `{"questions":[{"id":"education_level","question":"Age?","type":"choice","options":["12 клас"]}]}`, nil
	})
	resp, err := c.GenerateQuestions(context.Background(), model.QuestionsRequest{
		Category: "ученик", UserInput: "test", Lang: "bg",
	})
	if err != nil || resp.Questions[0].ID != "education_level" {
		t.Fatalf("err=%v resp=%#v", err, resp)
	}
}

func TestGenerateQuestions_FallbackAndBilling(t *testing.T) {
	c := testGeminiClient(func(ctx context.Context, system, user string, withSearch bool) (string, error) {
		return "", errors.New("gemini billing: quota exceeded prepayment")
	})
	_, err := c.GenerateQuestions(context.Background(), model.QuestionsRequest{
		Category: "ученик", UserInput: "test", Lang: "bg",
	})
	if err == nil || !IsBillingError(err) {
		t.Fatalf("got %v", err)
	}

	c2 := testGeminiClient(func(ctx context.Context, system, user string, withSearch bool) (string, error) {
		return "not-json", nil
	})
	resp, err := c2.GenerateQuestions(context.Background(), model.QuestionsRequest{
		Category: "ученик", UserInput: "test", Lang: "bg",
	})
	if err != nil || resp.Questions[0].ID != "education_level" {
		t.Fatalf("fallback failed: %v %#v", err, resp)
	}
}

func TestGenerateResults_TwoStepStub(t *testing.T) {
	step := 0
	c := testGeminiClient(func(ctx context.Context, system, user string, withSearch bool) (string, error) {
		step++
		if withSearch {
			return "research facts", nil
		}
		return `{"intro":"ok","items":[{"id":"p1","title":"TU","summary":"s","whyFits":"-","opportunities":"-","honestAssessment":"-","nextSteps":"-"}],"opportunities":[]}`, nil
	})
	resp, err := c.GenerateResults(context.Background(), model.ResultsRequest{
		Category: "ученик", UserInput: "test", Lang: "bg", Answers: map[string]string{},
	})
	if err != nil || step < 2 || len(resp.Items) != 1 {
		t.Fatalf("step=%d err=%v items=%d", step, err, len(resp.Items))
	}
}

func TestValidate_WithStub(t *testing.T) {
	c := testGeminiClient(func(ctx context.Context, system, user string, withSearch bool) (string, error) {
		return `{"status":"ok"}`, nil
	})
	if err := c.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}

	c2 := testGeminiClient(func(ctx context.Context, system, user string, withSearch bool) (string, error) {
		return "", errors.New("gemini 503 unavailable high demand")
	})
	err := c2.Validate(context.Background())
	if err == nil || !IsTransientError(err) {
		t.Fatalf("got %v", err)
	}
}

func TestGenerateQuestions_ValidationError(t *testing.T) {
	c := testGeminiClient(nil)
	_, err := c.GenerateQuestions(context.Background(), model.QuestionsRequest{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGenerateWithRetry_EmptyContentRetries(t *testing.T) {
	attempts := 0
	c := testGeminiClient(func(ctx context.Context, system, user string, withSearch bool) (string, error) {
		attempts++
		if attempts < 2 {
			return "", nil
		}
		return `{"status":"ok"}`, nil
	})
	text, err := c.generateWithRetry(context.Background(), "sys", "user", false)
	if err != nil || text == "" || attempts < 2 {
		t.Fatalf("attempts=%d err=%v text=%q", attempts, err, text)
	}
}

func TestResponseText_EmptyCandidates(t *testing.T) {
	if responseText(nil) != "" {
		t.Fatal("expected empty")
	}
	if responseText(&genai.GenerateContentResponse{}) != "" {
		t.Fatal("expected empty for no candidates")
	}
}

func TestResponseText_CollectsParts(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{{
			Content: &genai.Content{
				Parts: []*genai.Part{{Text: "hello"}, {Text: " world"}},
			},
		}},
	}
	if got := responseText(resp); got != "hello world" {
		t.Fatalf("got %q", got)
	}
}

func TestFinishReason(t *testing.T) {
	if finishReason(nil) != "no candidates" {
		t.Fatal("nil response")
	}
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{{FinishReason: genai.FinishReasonStop}},
	}
	if finishReason(resp) != "STOP" {
		t.Fatalf("got %q", finishReason(resp))
	}
}