package groq

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

func mockGroqServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(handler))
}

func jsonResponse(content string) []byte {
	b, _ := json.Marshal(chatResponse{
		Choices: []struct {
			Message chatMessage `json:"message"`
		}{{Message: chatMessage{Role: "assistant", Content: content}}},
	})
	return b
}

func testGroqClient(srv *httptest.Server) *Client {
	c := New(config.GroqConfig{
		APIKey:         "gsk_test",
		Model:          "llama-3.1-8b-instant",
		ResultsModel:   "groq/compound-mini",
		RequestTimeout: 5 * time.Second,
	})
	c.endpoint = srv.URL
	c.http = srv.Client()
	return c
}

func TestClient_Name(t *testing.T) {
	c := New(config.GroqConfig{APIKey: "gsk_test", Model: "llama-3.1-8b-instant"})
	if c.Name() != "groq" {
		t.Fatalf("name %q", c.Name())
	}
}

func TestGenerateQuestions_Success(t *testing.T) {
	srv := mockGroqServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(jsonResponse(`{"questions":[{"id":"education_level","question":"Age?","type":"choice","options":["12"]}]}`))
	})
	defer srv.Close()

	c := testGroqClient(srv)
	resp, err := c.GenerateQuestions(context.Background(), model.QuestionsRequest{
		Category: "ученик", UserInput: "test", Lang: "bg",
	})
	if err != nil || resp.Questions[0].ID != "education_level" {
		t.Fatalf("err=%v resp=%#v", err, resp)
	}
}

func TestGenerateResults_TwoStepPipeline(t *testing.T) {
	calls := 0
	srv := mockGroqServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Write(jsonResponse(`research report about universities`))
			return
		}
		w.Write(jsonResponse(`{"intro":"ok","items":[{"id":"p1","title":"PU — Sec","summary":"s","whyFits":"- f","opportunities":"o","honestAssessment":"- h","nextSteps":"- n"}],"opportunities":[{"title":"Hack","type":"Хакатон","description":"d","howToApply":"a","timing":"t"}]}`))
	})
	defer srv.Close()

	c := testGroqClient(srv)
	resp, err := c.GenerateResults(context.Background(), model.ResultsRequest{
		Category: "ученик", UserInput: "киберсигурност", Lang: "bg",
		Answers: map[string]string{"education_level": "12 клас"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls < 2 || len(resp.Items) != 1 {
		t.Fatalf("calls=%d items=%d", calls, len(resp.Items))
	}
}

func TestGenerateResults_RateLimitFallbackResearch(t *testing.T) {
	calls := 0
	srv := mockGroqServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"rate_limit_exceeded","code":"rate_limit_exceeded"}}`))
			return
		}
		if calls == 2 {
			w.Write(jsonResponse(`offline research`))
			return
		}
		w.Write(jsonResponse(`{"intro":"ok","items":[{"id":"p1","title":"Uni","summary":"s","whyFits":"-","opportunities":"-","honestAssessment":"-","nextSteps":"-"}],"opportunities":[]}`))
	})
	defer srv.Close()

	c := testGroqClient(srv)
	if _, err := c.GenerateResults(context.Background(), model.ResultsRequest{
		Category: "ученик", UserInput: "test", Lang: "bg", Answers: map[string]string{},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateResults_JSONRepair(t *testing.T) {
	calls := 0
	payload := `{"intro":"ok","items":[{"id":"p1","title":"Uni","summary":"s","whyFits":"-","opportunities":"-","honestAssessment":"-","nextSteps":"-"}],"opportunities":[]}`
	srv := mockGroqServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch calls {
		case 1:
			w.Write(jsonResponse(`research`))
		case 2:
			w.Write(jsonResponse(`not-json`))
		case 3:
			w.Write(jsonResponse(payload))
		}
	})
	defer srv.Close()
	c := testGroqClient(srv)
	if _, err := c.GenerateResults(context.Background(), model.ResultsRequest{
		Category: "ученик", UserInput: "test", Lang: "bg", Answers: map[string]string{},
	}); err != nil {
		t.Fatal(err)
	}
	if calls < 3 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestValidate_Success(t *testing.T) {
	srv := mockGroqServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(jsonResponse(`{"status":"ok"}`))
	})
	defer srv.Close()
	c := testGroqClient(srv)
	if err := c.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateQuestions_ValidationError(t *testing.T) {
	c := New(config.GroqConfig{APIKey: "gsk_test", Model: "m", ResultsModel: "m", RequestTimeout: time.Second})
	_, err := c.GenerateQuestions(context.Background(), model.QuestionsRequest{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidate_BillingError(t *testing.T) {
	srv := mockGroqServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write([]byte(`{"error":{"message":"insufficient quota","code":"billing"}}`))
	})
	defer srv.Close()
	c := testGroqClient(srv)
	err := c.Validate(context.Background())
	if err == nil || !IsBillingError(err) {
		t.Fatalf("got %v", err)
	}
}