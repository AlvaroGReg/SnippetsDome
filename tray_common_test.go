package main

import "testing"

func TestTraySnippetTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{name: "empty title", title: "  ", want: "Untitled snippet"},
		{name: "trimmed title", title: "  Go example  ", want: "Go example"},
		{name: "long title", title: "12345678901234567890123456789012345678901234567890", want: "12345678901234567890123456789012345678901234567…"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := traySnippetTitle(test.title); got != test.want {
				t.Errorf("traySnippetTitle(%q) = %q, want %q", test.title, got, test.want)
			}
		})
	}
}
