package guidanceai

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/atilatair/realput-bg/backend/internal/model"
)

func TestExtractJSON_FencedBlock(t *testing.T) {
	raw, err := ExtractJSON("Here you go:\n```json\n{\"ok\":true}\n```")
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]bool
	if err := json.Unmarshal(raw, &out); err != nil || !out["ok"] {
		t.Fatalf("unexpected payload: %s", raw)
	}
}

func TestExtractJSON_TrailingCommaRepair(t *testing.T) {
	raw, err := ExtractJSON(`{"items":[{"id":"a",},]}`)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(raw) {
		t.Fatalf("invalid json after repair: %s", raw)
	}
}

func TestExtractJSON_NoObject(t *testing.T) {
	if _, err := ExtractJSON("no json here"); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseResultsPayload_StringAndArrayFields(t *testing.T) {
	in := []byte(`{
		"intro": "hello",
		"items": [{
			"id": "p1",
			"title": "Uni — CS",
			"summary": "good",
			"whyFits": ["fits you", "strong math"],
			"opportunities": "long text",
			"honestAssessment": "- hard",
			"nextSteps": ["step one", "step two"]
		}],
		"opportunities": [{"title": "Hackathon", "type": "Събитие"}]
	}`)
	out, err := ParseResultsPayload(in)
	if err != nil {
		t.Fatal(err)
	}
	if out.Items[0].WhyFits == "" || !strings.Contains(out.Items[0].WhyFits, "- fits you") {
		t.Fatalf("whyFits not coerced: %q", out.Items[0].WhyFits)
	}
	if !strings.Contains(out.Items[0].NextSteps, "- step one") {
		t.Fatalf("nextSteps not coerced: %q", out.Items[0].NextSteps)
	}
}

func TestParseResultsPayload_SanitizesURLs(t *testing.T) {
	in := []byte(`{
		"intro": "see https://btu.bg?utm_source=openai",
		"items": [],
		"opportunities": []
	}`)
	out, err := ParseResultsPayload(in)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.Intro, "utm_source") {
		t.Fatalf("utm not stripped: %q", out.Intro)
	}
}

func TestNormalizeResults_NilSlices(t *testing.T) {
	p := &model.ResultsPayload{}
	NormalizeResults(p)
	if p.Items == nil || p.Opportunities == nil {
		t.Fatal("expected empty slices")
	}
}

func TestFallbackQuestions_HasEducationFirst_BG(t *testing.T) {
	resp := FallbackQuestions("bg")
	if resp.Questions[0].ID != "education_level" {
		t.Fatalf("got %s", resp.Questions[0].ID)
	}
}

func TestFallbackQuestions_HasEducationFirst_EN(t *testing.T) {
	resp := FallbackQuestions("en")
	if resp.Questions[0].ID != "education_level" {
		t.Fatalf("got %s", resp.Questions[0].ID)
	}
	if !strings.Contains(resp.Questions[0].Question, "old") {
		t.Fatalf("expected english question: %q", resp.Questions[0].Question)
	}
}

func TestNormalizeQuestions_FillsNilOptions(t *testing.T) {
	resp := &model.QuestionsResponse{
		Questions: []model.Question{{ID: "x", Type: "text", Options: nil}},
	}
	NormalizeQuestions(resp, "bg")
	if resp.Questions[0].Options == nil {
		t.Fatal("options should not be nil")
	}
}