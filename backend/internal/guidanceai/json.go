package guidanceai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/atilatair/realput-bg/backend/internal/model"
)

var fencedJSON = regexp.MustCompile("(?s)```(?:json)?\\s*([\\s\\S]*?)```")

// ExtractJSON pulls the first JSON object from model text, including fenced blocks.
func ExtractJSON(text string) ([]byte, error) {
	raw := strings.TrimSpace(text)
	if match := fencedJSON.FindStringSubmatch(raw); len(match) > 1 {
		raw = strings.TrimSpace(match[1])
	}

	slice, ok := extractJSONObject(raw)
	if !ok {
		return nil, fmt.Errorf("no json object found in model output")
	}

	if json.Valid(slice) {
		return slice, nil
	}

	repaired := repairJSON(slice)
	if json.Valid(repaired) {
		return repaired, nil
	}
	return nil, fmt.Errorf("invalid json in model output")
}

func extractJSONObject(s string) ([]byte, bool) {
	start := strings.Index(s, "{")
	if start == -1 {
		return nil, false
	}

	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			switch ch {
			case '\\':
				escaped = true
			case '"':
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return []byte(s[start : i+1]), true
			}
		}
	}
	return nil, false
}

var trailingComma = regexp.MustCompile(`,\s*([}\]])`)

func repairJSON(raw []byte) []byte {
	out := trailingComma.ReplaceAll(raw, []byte("$1"))
	return bytes.TrimSpace(out)
}

func coerceMarkdownField(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}
	var items []string
	if err := json.Unmarshal(raw, &items); err == nil {
		if len(items) == 0 {
			return ""
		}
		var b strings.Builder
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			if b.Len() > 0 {
				b.WriteByte('\n')
			}
			if strings.HasPrefix(item, "- ") || strings.HasPrefix(item, "* ") {
				b.WriteString(item)
			} else {
				fmt.Fprintf(&b, "- %s", item)
			}
		}
		return b.String()
	}
	return strings.TrimSpace(string(raw))
}

type rawResultsPayload struct {
	Intro         string              `json:"intro"`
	Items         []rawResultItem     `json:"items"`
	Opportunities []model.Opportunity `json:"opportunities"`
}

type rawResultItem struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	Summary          string          `json:"summary"`
	WhyFits          json.RawMessage `json:"whyFits"`
	Opportunities    json.RawMessage `json:"opportunities"`
	HonestAssessment json.RawMessage `json:"honestAssessment"`
	NextSteps        json.RawMessage `json:"nextSteps"`
}

// ParseResultsPayload decodes flexible model JSON into ResultsPayload.
func ParseResultsPayload(raw []byte) (*model.ResultsPayload, error) {
	var parsed rawResultsPayload
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse results json: %w", err)
	}

	out := &model.ResultsPayload{
		Intro:         parsed.Intro,
		Opportunities: parsed.Opportunities,
		Items:         make([]model.ResultItem, 0, len(parsed.Items)),
	}
	for _, item := range parsed.Items {
		out.Items = append(out.Items, model.ResultItem{
			ID:               item.ID,
			Title:            item.Title,
			Summary:          item.Summary,
			WhyFits:          coerceMarkdownField(item.WhyFits),
			Opportunities:    coerceMarkdownField(item.Opportunities),
			HonestAssessment: coerceMarkdownField(item.HonestAssessment),
			NextSteps:        coerceMarkdownField(item.NextSteps),
		})
	}
	NormalizeResults(out)
	sanitizeResultsPayload(out)
	return out, nil
}

// NormalizeQuestions ensures question options are never nil and education_level is first.
func NormalizeQuestions(resp *model.QuestionsResponse, lang string) {
	for i := range resp.Questions {
		if resp.Questions[i].Options == nil {
			resp.Questions[i].Options = []string{}
		}
	}
	ensureEducationLevelFirst(resp, lang)
}

func ensureEducationLevelFirst(resp *model.QuestionsResponse, lang string) {
	defaultQ := DefaultEducationLevelQuestion(lang)

	var found *model.Question
	rest := make([]model.Question, 0, len(resp.Questions))
	for i := range resp.Questions {
		if resp.Questions[i].ID == "education_level" {
			q := resp.Questions[i]
			if q.Type == "" {
				q.Type = "choice"
			}
			if len(q.Options) == 0 {
				q.Options = defaultQ.Options
			}
			found = &q
			continue
		}
		rest = append(rest, resp.Questions[i])
	}

	if found != nil {
		resp.Questions = append([]model.Question{*found}, rest...)
		return
	}

	resp.Questions = append([]model.Question{defaultQ}, rest...)
}

// NormalizeResults ensures result slices are never nil.
func NormalizeResults(payload *model.ResultsPayload) {
	if payload.Items == nil {
		payload.Items = []model.ResultItem{}
	}
	if payload.Opportunities == nil {
		payload.Opportunities = []model.Opportunity{}
	}
}

// FallbackQuestions returns generic questions when the model fails.
func FallbackQuestions(lang string) *model.QuestionsResponse {
	resp := &model.QuestionsResponse{
		Questions: []model.Question{
			{
				ID:       "goal",
				Question: "Какво е най-важно за теб в момента?",
				Type:     "choice",
				Options:  []string{"Да избера специалност", "Да намеря стаж/работа", "Да разбера възможностите си", "Друго"},
			},
			{
				ID:       "interest",
				Question: "Кои области или предмети те вълнуват най-много?",
				Type:     "text",
				Options:  []string{},
			},
			{
				ID:       "location",
				Question: "В кой град или регион би искал да учиш, работиш или се развиваш?",
				Type:     "choice",
				Options:  []string{"София", "Пловдив", "Варна", "Бургас", "Друг град в България", "Цяла България / без значение"},
			},
			{
				ID:       "timeline",
				Question: "В какъв период искаш да предприемеш следваща стъпка?",
				Type:     "choice",
				Options:  []string{"Веднага", "Тази година", "Следващата година", "Все още не знам"},
			},
		},
	}
	NormalizeQuestions(resp, lang)
	return resp
}