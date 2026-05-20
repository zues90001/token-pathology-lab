package treatment

import (
	"strings"
	"testing"
)

// Helpers tested here are pure schema validation; no MiMo network calls.

func TestEntryIDSlugStable(t *testing.T) {
	cases := []struct {
		title    string
		expected string
	}{
		{"Excessive JSON wrapping in MiMo V2.5", "excessive-json-wrapping-in-mimo-v2-5"},
		{"Reasoning  leak / output", "reasoning-leak-output"},
		{"   leading & trailing   ", "leading-trailing"},
	}
	for _, c := range cases {
		got := slugify(c.title)
		if got != c.expected {
			t.Errorf("slugify(%q) = %q, want %q", c.title, got, c.expected)
		}
	}
}

func TestValidateEntryRejectsMissingFields(t *testing.T) {
	bad := AtlasEntry{Title: ""}
	if err := ValidateEntry(&bad); err == nil {
		t.Fatal("expected error for empty title")
	} else if !strings.Contains(err.Error(), "title") {
		t.Errorf("expected title error, got %v", err)
	}
}

func TestValidateEntryAcceptsMinimal(t *testing.T) {
	ok := AtlasEntry{
		Title:        "Sample diagnosis",
		Category:     "reasoning-leak",
		BeforePrompt: "...",
		AfterPrompt:  "...",
		TokenDelta:   500,
	}
	if err := ValidateEntry(&ok); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}
