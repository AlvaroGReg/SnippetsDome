package main

import (
	"fmt"
	"strings"

	"SnippetsDome/internal/domain"
)

func traySnippets(app *App) []domain.Snippet {
	snippets, err := app.snippets.List()
	if err != nil {
		return nil
	}

	limit := app.snippets.TraySnippetLimit()
	if len(snippets) > limit {
		snippets = snippets[:limit]
	}
	return append([]domain.Snippet(nil), snippets...)
}

func traySnippetTitle(title string) string {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return "Untitled snippet"
	}
	const maximumLength = 48
	if len(trimmed) > maximumLength {
		return fmt.Sprintf("%s…", trimmed[:maximumLength-1])
	}
	return trimmed
}
