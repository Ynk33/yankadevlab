package handler

import (
	"net/url"
	"strings"
)

// safeRedirect returns rd when it is an https URL on cookieDomain or one of its subdomains, fallback otherwise.
func safeRedirect(rd, cookieDomain, fallback string) string {
	u, err := url.Parse(rd)
	if err != nil || u.Scheme != "https" || u.User != nil {
		return fallback
	}

	domain := strings.TrimPrefix(cookieDomain, ".")
	host := u.Hostname()
	if host != domain && !strings.HasSuffix(host, "."+domain) {
		return fallback
	}
	return u.String()
}
