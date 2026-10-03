package handler

import "testing"

func TestSafeRedirect(t *testing.T) {
	const fallback = "https://dashboard.example.tech"

	tests := []struct {
		name string
		rd   string
		want string
	}{
		{name: "subdomain", rd: "https://app.example.tech/path?q=1", want: "https://app.example.tech/path?q=1"},
		{name: "apex", rd: "https://example.tech/", want: "https://example.tech/"},
		{name: "http scheme", rd: "http://app.example.tech/", want: fallback},
		{name: "lookalike domain", rd: "https://evil-example.tech/", want: fallback},
		{name: "suffix attack", rd: "https://example.tech.evil.com/", want: fallback},
		{name: "userinfo", rd: "https://example.tech@evil.com/", want: fallback},
		{name: "relative", rd: "/path", want: fallback},
		{name: "protocol relative", rd: "//evil.com", want: fallback},
		{name: "javascript", rd: "javascript:alert(1)", want: fallback},
		{name: "empty", rd: "", want: fallback},
		{name: "garbage", rd: "%zz", want: fallback},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := safeRedirect(tt.rd, ".example.tech", fallback); got != tt.want {
				t.Errorf("safeRedirect(%q) = %q, want %q", tt.rd, got, tt.want)
			}
		})
	}
}
