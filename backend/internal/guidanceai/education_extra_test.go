package guidanceai

import "testing"

func TestIsUniversityEducationLevel(t *testing.T) {
	if !isUniversityEducationLevel("17–19 г. (12 клас или завършил) — търся университет") {
		t.Fatal("expected university level")
	}
	if isUniversityEducationLevel("12–13 г. (5–7 клас)") {
		t.Fatal("expected high school level")
	}
}

func TestIsHighSchoolEducationLevel(t *testing.T) {
	if !isHighSchoolEducationLevel("12–13 г. (5–7 клас) — избор на гимназия") {
		t.Fatal("expected true")
	}
}

func TestExtractAge(t *testing.T) {
	if extractAge("На 17 години съм") != 17 {
		t.Fatal("expected 17")
	}
	if extractAge("no age") != 0 {
		t.Fatal("expected 0")
	}
}

func TestExtractAnswerValue(t *testing.T) {
	text := "- education_level: 12 клас\n- location: Пловдив"
	if got := extractAnswerValue(text, "education_level"); got != "12 клас" {
		t.Fatalf("got %q", got)
	}
	if got := extractAnswerValue(text, "missing"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestHasHighSchoolOnlySignals(t *testing.T) {
	if !hasHighSchoolOnlySignals("След 7 клас търся гимназия") {
		t.Fatal("expected high school signals")
	}
	if hasHighSchoolOnlySignals("На 18 искам университет") {
		t.Fatal("expected no high school-only signals")
	}
}