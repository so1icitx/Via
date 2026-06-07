package handler

import (
	"net/url"
	"strings"
)

// resolveRedirectTarget returns a safe post-auth redirect limited to the frontend origin.
func resolveRedirectTarget(frontendURL, target string) string {
	frontendURL = strings.TrimRight(strings.TrimSpace(frontendURL), "/")
	target = strings.TrimSpace(target)
	if target == "" {
		return frontendURL
	}
	if strings.HasPrefix(target, "/") {
		return frontendURL + target
	}

	base, err := url.Parse(frontendURL)
	if err != nil {
		return frontendURL
	}
	u, err := url.Parse(target)
	if err != nil {
		return frontendURL
	}
	if u.Scheme == base.Scheme && u.Host == base.Host {
		return target
	}
	return frontendURL
}