package guidanceai

import (
	"strings"
	"testing"
)

func TestSanitizeModelText_OpenAICitation(t *testing.T) {
	in := "Регистрация на https://www.codeburgas.com ([invest.burgas.bg](https://invest.burgas.bg/bg/uchenicheskoto-sastezanie?utm_source=openai))"
	want := "Регистрация на [invest.burgas.bg](https://invest.burgas.bg/bg/uchenicheskoto-sastezanie)"
	got := SanitizeModelText(in)
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSanitizeModelText_StandaloneCitation(t *testing.T) {
	in := "Виж ([codingburgas.bg](https://codingburgas.bg/admission?utm_source=openai&utm_medium=web))"
	want := "Виж [codingburgas.bg](https://codingburgas.bg/admission)"
	got := SanitizeModelText(in)
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSanitizeModelText_MessyMarkdownLink(t *testing.T) {
	in := "Посетете [pmgbs.com/прием-viii-клас](http://pmgbs.com/%D0%BF%D1%80%D0%B8%D0%B5%D0%BC-viii-%D0%BA%D0%BB%D0%B0%D1%81)"
	got := SanitizeModelText(in)
	if !strings.Contains(got, "https://pmgbs.com") {
		t.Fatalf("got %q, want https pmgbs url", got)
	}
	if strings.Contains(got, "%D0") {
		t.Fatalf("label should not keep percent-encoding noise: %q", got)
	}
}

func TestSanitizeModelText_BareURL(t *testing.T) {
	in := "Сайт: https://btu.bg/page?utm_source=openai"
	want := "Сайт: [btu.bg/page](https://btu.bg/page)"
	got := SanitizeModelText(in)
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}