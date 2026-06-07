package guidanceai

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/atilatair/realput-bg/backend/internal/model"
)

const preferHTTPS = true

var (
	openAICitation = regexp.MustCompile(`\s*\(\[([^\]]+)\]\(([^)]+)\)\)`)
	urlThenCitation = regexp.MustCompile(`(https?://[^\s\])]+)\s*\(\[([^\]]+)\]\(([^)]+)\)\)`)
	markdownLink = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)
	bareURL = regexp.MustCompile(`https?://[^\s\])<>"']+`)
)

func SanitizeModelText(s string) string {
	if s == "" {
		return s
	}

	out := s
	out = urlThenCitation.ReplaceAllStringFunc(out, func(match string) string {
		parts := urlThenCitation.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		label := strings.TrimSpace(parts[2])
		citeURL := cleanURL(parts[3])
		if citeURL == "" {
			return cleanURL(parts[1])
		}
		if label == "" {
			return citeURL
		}
		return "[" + label + "](" + citeURL + ")"
	})

	out = openAICitation.ReplaceAllStringFunc(out, func(match string) string {
		parts := openAICitation.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		label := strings.TrimSpace(parts[1])
		citeURL := cleanURL(parts[2])
		if citeURL == "" {
			return ""
		}
		if label == "" {
			return citeURL
		}
		prefix := ""
		if strings.TrimLeft(match, " \t") != match {
			prefix = " "
		}
		return prefix + "[" + label + "](" + citeURL + ")"
	})

	out = markdownLink.ReplaceAllStringFunc(out, func(match string) string {
		parts := markdownLink.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		cleaned := cleanURL(parts[2])
		if cleaned == "" {
			return parts[1]
		}
		return "[" + cleanLinkLabel(parts[1], cleaned) + "](" + cleaned + ")"
	})

	out = autolinkBareURLs(out)
	return strings.TrimSpace(out)
}

func autolinkBareURLs(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		loc := bareURL.FindStringIndex(s[i:])
		if loc == nil {
			b.WriteString(s[i:])
			break
		}
		start := i + loc[0]
		end := i + loc[1]
		b.WriteString(s[i:start])

		raw := s[start:end]
		if isInsideMarkdownLink(s, start) {
			b.WriteString(cleanURL(raw))
		} else {
			cleaned := cleanURL(raw)
			if cleaned == "" {
				b.WriteString(raw)
			} else {
				label := linkLabel(cleaned)
				b.WriteString("[" + label + "](" + cleaned + ")")
			}
		}
		i = end
	}
	return b.String()
}

func isInsideMarkdownLink(s string, urlStart int) bool {
	before := s[:urlStart]
	open := strings.LastIndex(before, "](")
	if open == -1 {
		return false
	}
	close := strings.Index(s[urlStart:], ")")
	if close == -1 {
		return false
	}
	segment := s[open+2 : urlStart+close]
	return strings.HasPrefix(segment, "http://") || strings.HasPrefix(segment, "https://")
}

func linkLabel(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return rawURL
	}
	host := strings.TrimPrefix(parsed.Host, "www.")
	if parsed.Path != "" && parsed.Path != "/" {
		if decoded, err := url.PathUnescape(parsed.EscapedPath()); err == nil && decoded != "" {
			return host + decoded
		}
		return host + parsed.Path
	}
	return host
}

func cleanLinkLabel(label, cleanedURL string) string {
	label = strings.TrimSpace(label)
	defaultLabel := linkLabel(cleanedURL)
	if label == "" {
		return defaultLabel
	}
	if strings.Contains(label, "%") || strings.ContainsRune(label, '/') && strings.Contains(label, "://") {
		return defaultLabel
	}
	// Prefer short host label when model invents noisy paths in the link text.
	if len(label) > 48 {
		return defaultLabel
	}
	return label
}

func cleanURL(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimRight(raw, ".,;:!?)")
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return raw
	}
	if preferHTTPS && parsed.Scheme == "http" {
		parsed.Scheme = "https"
	}

	q := parsed.Query()
	for key := range q {
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "utm_") || lower == "openai" {
			q.Del(key)
		}
	}
	parsed.RawQuery = q.Encode()
	return readableURL(parsed)
}

func readableURL(parsed *url.URL) string {
	if parsed == nil {
		return ""
	}
	path := parsed.EscapedPath()
	if decoded, err := url.PathUnescape(path); err == nil && decoded != "" {
		path = decoded
	}
	if path == "" {
		path = "/"
	}
	out := parsed.Scheme + "://" + parsed.Host + path
	if parsed.RawQuery != "" {
		out += "?" + parsed.RawQuery
	}
	return strings.TrimRight(out, "/")
}

func sanitizeResultsPayload(payload *model.ResultsPayload) {
	if payload == nil {
		return
	}
	payload.Intro = SanitizeModelText(payload.Intro)
	for i := range payload.Items {
		item := &payload.Items[i]
		item.Title = SanitizeModelText(item.Title)
		item.Summary = SanitizeModelText(item.Summary)
		item.WhyFits = SanitizeModelText(item.WhyFits)
		item.Opportunities = SanitizeModelText(item.Opportunities)
		item.HonestAssessment = SanitizeModelText(item.HonestAssessment)
		item.NextSteps = SanitizeModelText(item.NextSteps)
	}
	for i := range payload.Opportunities {
		op := &payload.Opportunities[i]
		op.Title = SanitizeModelText(op.Title)
		op.Type = SanitizeModelText(op.Type)
		op.Description = SanitizeModelText(op.Description)
		op.HowToApply = SanitizeModelText(op.HowToApply)
		op.Timing = SanitizeModelText(op.Timing)
	}
}