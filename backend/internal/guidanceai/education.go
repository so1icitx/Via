package guidanceai

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/atilatair/realput-bg/backend/internal/model"
)

var agePattern = regexp.MustCompile(`(?i)(?:на\s+)?(\d{1,2})\s*(?:години|г\.|years?\s*old)`)

// DefaultEducationLevelQuestion returns the mandatory first question (age/class).
func DefaultEducationLevelQuestion(lang string) model.Question {
	if lang == "en" {
		return model.Question{
			ID:       "education_level",
			Question: "How old are you, or what grade are you in?",
			Type:     "choice",
			Options: []string{
				"12–13 (grades 5–7) — choosing a high school next",
				"14–15 (grades 8–9) — already in high school",
				"16–17 (grades 10–11)",
				"17–19 (grade 12 or graduated) — applying to university",
				"20+ — work, career change, or retraining",
			},
		}
	}
	return model.Question{
		ID:       "education_level",
		Question: "Колко години си / в кой клас си?",
		Type:     "choice",
		Options: []string{
			"12–13 г. (5–7 клас) — предстои избор на гимназия",
			"14–15 г. (8–9 клас) — вече съм в гимназия",
			"16–17 г. (10–11 клас)",
			"17–19 г. (12 клас или завършил) — търся университет",
			"20+ г. — работа, кариера или смяна на посока",
		},
	}
}

func hasUniversityIntent(text string) bool {
	signals := []string{
		"университет", "уни ", " uni", "бакалавър", "магистратура", "факултет",
		"колеж", "ву ", "ву/", "higher education", "university", "college",
		"бакалавърска", "кандидатствам за ву", "кандидатствам за университет",
		"12 клас или завършил", "завършил) — търся университет",
		"grade 12 or graduated", "applying to university",
	}
	for _, s := range signals {
		if strings.Contains(text, s) {
			return true
		}
	}

	if strings.Contains(text, "education_level:") || strings.Contains(text, "- education_level:") {
		level := extractAnswerValue(text, "education_level")
		if isUniversityEducationLevel(level) {
			return true
		}
	}

	if age := extractAge(text); age >= 17 && age <= 19 {
		if !hasHighSchoolOnlySignals(text) {
			return true
		}
	}

	if strings.Contains(text, "12 клас") || strings.Contains(text, "12-ти") ||
		strings.Contains(text, "завършил") || strings.Contains(text, "grade 12") {
		return !hasHighSchoolOnlySignals(text)
	}

	return false
}

func hasHighSchoolOnlySignals(text string) bool {
	signals := []string{
		"7 клас", "7-ми", "vii клас", "след 7", "viii клас", "viii",
		"избор на гимназия", "прием viii", "5–7 клас", "5-7 клас",
		"grades 5–7", "choosing a high school",
	}
	for _, s := range signals {
		if strings.Contains(text, s) {
			return true
		}
	}
	level := extractAnswerValue(text, "education_level")
	return isHighSchoolEducationLevel(level)
}

func isUniversityEducationLevel(level string) bool {
	level = strings.ToLower(level)
	return strings.Contains(level, "12 клас") ||
		strings.Contains(level, "завършил") ||
		strings.Contains(level, "университет") ||
		strings.Contains(level, "17–19") ||
		strings.Contains(level, "17-19") ||
		strings.Contains(level, "20+") ||
		strings.Contains(level, "grade 12") ||
		strings.Contains(level, "university") ||
		strings.Contains(level, "20+")
}

func isHighSchoolEducationLevel(level string) bool {
	level = strings.ToLower(level)
	return strings.Contains(level, "7 клас") ||
		strings.Contains(level, "5–7") ||
		strings.Contains(level, "5-7") ||
		strings.Contains(level, "избор на гимназия") ||
		strings.Contains(level, "8–9") ||
		strings.Contains(level, "8-9") ||
		strings.Contains(level, "grades 5–7") ||
		strings.Contains(level, "choosing a high school")
}

func extractAnswerValue(text, key string) string {
	keyLower := strings.ToLower(key)
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmedLower := strings.ToLower(trimmed)
		for _, prefix := range []string{"- " + keyLower + ":", keyLower + ":"} {
			if strings.HasPrefix(trimmedLower, prefix) {
				return strings.TrimSpace(trimmed[len(prefix):])
			}
		}
	}
	return ""
}

func extractAge(text string) int {
	m := agePattern.FindStringSubmatch(text)
	if len(m) < 2 {
		return 0
	}
	age, err := strconv.Atoi(m[1])
	if err != nil || age < 10 || age > 25 {
		return 0
	}
	return age
}

func universitySelectionGuidance() string {
	return `═══ ИЗРИЧНА ЦЕЛ: УНИВЕРСИТЕТ / КОЛЕЖ (БАКАЛАВЪР) ═══
Потребителят е в 12 клас, завършил гимназия или е на 17–19 г. и търси ВИСШЕ ОБРАЗОВАНИЕ.

items[] — ЗАДЪЛЖИТЕЛНО:
- МИНИМУМ 3 различни РЕАЛНИ университета или колежа с КОНКРЕТНИ бакалавърски специалности
- Пълно досие: прием 2026, ДЗИ, мин/макс бал, такси EUR, учебен план, стажове

ЗАБРАНЕНО в items[]:
- Гимназии, ПГ за прием след 7 клас, VIII клас
- Scratch, Robopartans, детски курсове
- Общи „курс по програмиране" без институция

Търси: „прием бакалавър [специалност] [град] 2026", „[университет] киберсигурност", „ДЗИ изисквания [специалност]".`
}