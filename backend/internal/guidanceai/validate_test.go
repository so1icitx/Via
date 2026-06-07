package guidanceai

import (
	"strings"
	"testing"

	"github.com/atilatair/realput-bg/backend/internal/model"
)

func TestValidateQuestionsRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     model.QuestionsRequest
		wantErr string
	}{
		{name: "ok", req: model.QuestionsRequest{Category: "ученик", UserInput: "искам ВУ"}},
		{name: "missing category", req: model.QuestionsRequest{UserInput: "x"}, wantErr: "category"},
		{name: "missing input", req: model.QuestionsRequest{Category: "ученик"}, wantErr: "userInput"},
		{name: "too long", req: model.QuestionsRequest{Category: "ученик", UserInput: strings.Repeat("а", 4001)}, wantErr: "too long"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateQuestionsRequest(tc.req)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
				t.Fatalf("got err=%v want contains %q", err, tc.wantErr)
			}
		})
	}
}

func TestValidateResultsRequest(t *testing.T) {
	err := ValidateResultsRequest(model.ResultsRequest{Category: "друго", UserInput: "test"})
	if err != nil {
		t.Fatal(err)
	}
	err = ValidateResultsRequest(model.ResultsRequest{Category: "", UserInput: "test"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}