package guidanceai

import (
	"strings"
	"testing"

	"github.com/atilatair/realput-bg/backend/internal/model"
)

func TestHasUniversityIntent_FromAgeAndInput(t *testing.T) {
	input := "На 17 години съм, интересувам се от мрежова сигурност и Linux."
	answers := "- education_level: 12 клас\n- location: Пловдив"
	combined := strings.ToLower(input + "\n" + answers)

	if !hasUniversityIntent(combined) {
		t.Fatal("expected university intent for 17yo with 12th grade answer")
	}
}

func TestHasUniversityIntent_NotHighSchoolPython(t *testing.T) {
	input := "На 17 години съм, интересувам се от Python и Linux. Искам университет."
	if !hasUniversityIntent(strings.ToLower(input)) {
		t.Fatal("expected university intent when user explicitly wants university")
	}
}

func TestHasHighSchoolIntent_After7thGrade(t *testing.T) {
	input := "Синът ми завършва 7 клас и търси гимназия с програмиране."
	answers := "- education_level: 12–13 г. (5–7 клас) — предстои избор на гимназия"
	combined := strings.ToLower(input + "\n" + answers)

	if hasUniversityIntent(combined) {
		t.Fatal("expected high school path, not university")
	}
	guidance := educationLevelFromAnswers(input, answers)
	if !strings.Contains(guidance, "ГИМНАЗИЯ") && !strings.Contains(guidance, "VIII") {
		t.Fatalf("expected high school guidance, got: %s", guidance)
	}
}

func TestEducationLevelFromAnswers_PrefersUniversity(t *testing.T) {
	input := "Искам да уча киберсигурност в Пловдив."
	answers := "- education_level: 17–19 г. (12 клас или завършил) — търся университет\n- location: Пловдив"

	guidance := educationLevelFromAnswers(input, answers)
	if !strings.Contains(guidance, "УНИВЕРСИТЕТ") {
		t.Fatalf("expected university guidance, got: %s", guidance)
	}
}

func TestEnsureEducationLevelFirst_InsertsWhenMissing(t *testing.T) {
	resp := &model.QuestionsResponse{
		Questions: []model.Question{
			{ID: "location", Question: "Къде?", Type: "choice", Options: []string{"София"}},
		},
	}
	ensureEducationLevelFirst(resp, "bg")
	if resp.Questions[0].ID != "education_level" {
		t.Fatalf("expected education_level first, got %s", resp.Questions[0].ID)
	}
}

func TestEnsureEducationLevelFirst_MovesWhenNotFirst(t *testing.T) {
	resp := &model.QuestionsResponse{
		Questions: []model.Question{
			{ID: "interest", Question: "Интереси?", Type: "text"},
			{ID: "education_level", Question: "Клас?", Type: "choice", Options: []string{"12 клас"}},
		},
	}
	ensureEducationLevelFirst(resp, "bg")
	if resp.Questions[0].ID != "education_level" {
		t.Fatalf("expected education_level first, got %s", resp.Questions[0].ID)
	}
}

func TestFallbackQuestions_EducationLevelFirst(t *testing.T) {
	resp := FallbackQuestions("bg")
	if len(resp.Questions) == 0 || resp.Questions[0].ID != "education_level" {
		t.Fatalf("expected education_level first in fallback, got %#v", resp.Questions)
	}
}