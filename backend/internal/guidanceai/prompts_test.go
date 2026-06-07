package guidanceai

import (
	"strings"
	"testing"
)

func TestLangInstruction(t *testing.T) {
	if !strings.Contains(LangInstruction("bg", false), "български") {
		t.Fatal("expected bulgarian instruction")
	}
	if !strings.Contains(LangInstruction("en", true), "English") {
		t.Fatal("expected english instruction")
	}
}

func TestQuestionsSystem_RequiresEducationLevelFirst(t *testing.T) {
	sys := QuestionsSystem("bg")
	if !strings.Contains(sys, `id: "education_level"`) {
		t.Fatal("missing education_level id")
	}
	if !strings.Contains(sys, "questions[0]") {
		t.Fatal("missing first-question rule")
	}
}

func TestResultsUser_UniversityPriority(t *testing.T) {
	user := ResultsUser("ученик",
		"На 17 години съм, искам киберсигурност в Пловдив",
		"- education_level: 17–19 г. (12 клас или завършил) — търся университет\n- location: Пловдив",
	)
	if !strings.Contains(user, "УНИВЕРСИТЕТ") {
		t.Fatalf("expected university guidance in prompt: %s", user)
	}
}

func TestResearchUser_HighSchoolAfter7th(t *testing.T) {
	user := ResearchUser("родител",
		"Синът завършва 7 клас",
		"- education_level: 12–13 г. (5–7 клас) — предстои избор на гимназия",
	)
	if !strings.Contains(user, "VIII") && !strings.Contains(user, "ГИМНАЗИЯ") {
		t.Fatalf("expected high school guidance: %s", user)
	}
}

func TestIntentFromContext_UniversityWinsOverPython(t *testing.T) {
	got := intentFromContext("На 17 години, Python, искам университет", "")
	if !strings.Contains(got, "УНИВЕРСИТЕТ") {
		t.Fatalf("got %q", got)
	}
}

func TestResultsSystem_ContainsDateAwareness(t *testing.T) {
	sys := ResultsSystem("bg")
	if !strings.Contains(sys, "JSON") || !strings.Contains(sys, "items") {
		t.Fatal("expected structured output instructions")
	}
}

func TestResearchUser_IncludesEducationLevel(t *testing.T) {
	user := ResearchUser("ученик", "test", "- education_level: 12 клас")
	if !strings.Contains(user, "12 клас") {
		t.Fatal("expected answers in prompt")
	}
}

func TestEducationLevelFromAnswers_DefaultEnglish(t *testing.T) {
	q := DefaultEducationLevelQuestion("en")
	if q.ID != "education_level" || len(q.Options) < 4 {
		t.Fatalf("unexpected question: %#v", q)
	}
}

func TestQuestionsUser_IncludesCategory(t *testing.T) {
	user := QuestionsUser("ученик", "искам IT")
	if !strings.Contains(user, "ученик") || !strings.Contains(user, "искам IT") {
		t.Fatalf("unexpected prompt: %s", user)
	}
}

func TestResearchSystem_ContainsSearchGuidance(t *testing.T) {
	sys := ResearchSystem("bg")
	if !strings.Contains(sys, "търсене") || !strings.Contains(sys, "JSON") {
		t.Fatal("expected research instructions")
	}
}

func TestResultsFromResearchUser_IncludesResearch(t *testing.T) {
	user := ResultsFromResearchUser("ученик", "input", "- education_level: 12 клас", "research facts")
	if !strings.Contains(user, "research facts") || !strings.Contains(user, "12 клас") {
		t.Fatalf("unexpected prompt: %s", user)
	}
}

func TestEducationLevelFromAnswers_University(t *testing.T) {
	got := educationLevelFromAnswers("университет", "- education_level: 17–19 г. (12 клас или завършил) — търся университет")
	if !strings.Contains(got, "УНИВЕРСИТЕТ") {
		t.Fatalf("got %q", got)
	}
}