package guidanceai

import "testing"

func TestSanitizeModelText_Empty(t *testing.T) {
	if SanitizeModelText("") != "" {
		t.Fatal("expected empty")
	}
}

func TestCleanURL_HTTPToHTTPS(t *testing.T) {
	got := cleanURL("http://example.com/path?utm_source=openai")
	if got != "https://example.com/path" {
		t.Fatalf("got %q", got)
	}
}

func TestLinkLabel_WithPath(t *testing.T) {
	got := linkLabel("https://btu.bg/admission")
	if got == "" {
		t.Fatal("expected label")
	}
}

func TestCleanLinkLabel_LongLabelUsesHost(t *testing.T) {
	long := "this-is-a-very-long-label-that-should-fallback-to-host-name-default"
	got := cleanLinkLabel(long, "https://btu.bg/admission")
	if got == long {
		t.Fatal("expected shortened label")
	}
}

func TestSanitizeResultsPayload_NilSafe(t *testing.T) {
	sanitizeResultsPayload(nil)
}