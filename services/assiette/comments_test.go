package main

import (
	"strings"
	"testing"
)

func TestCommentBody(t *testing.T) {
	tests := []struct {
		name, in, want string
		ok             bool
	}{
		{"plain", "Ajouter du citron", "Ajouter du citron", true},
		{"trimmed", "  \n un peu moins de sel \t", "un peu moins de sel", true},
		{"empty", "", "", false},
		{"blank", " \n\t ", "", false},
		{"max length in runes", strings.Repeat("é", maxCommentRunes), strings.Repeat("é", maxCommentRunes), true},
		{"too long", strings.Repeat("a", maxCommentRunes+1), strings.Repeat("a", maxCommentRunes+1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := commentBody(tt.in)
			if got != tt.want || ok != tt.ok {
				t.Errorf("got (%q, %v), want (%q, %v)", got, ok, tt.want, tt.ok)
			}
		})
	}
}
