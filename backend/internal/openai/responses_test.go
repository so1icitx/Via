package openai

import "testing"

func TestExtractResponsesText_OutputText(t *testing.T) {
	got := extractResponsesText(responsesResponse{OutputText: " hello "})
	if got != "hello" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractResponsesText_OutputItems(t *testing.T) {
	got := extractResponsesText(responsesResponse{
		Output: []responsesOutputItem{{
			Type: "message",
			Content: []responsesContentPart{
				{Type: "output_text", Text: "line one"},
				{Type: "output_text", Text: "line two"},
			},
		}},
	})
	if got != "line one\nline two" {
		t.Fatalf("got %q", got)
	}
}