package domain

const DefaultTraySnippetLimit = 5

const (
	LanguageEnglish = "en"
	LanguageSpanish = "es"
)

// AppConfig contains the preferences that are persisted between application runs.
type AppConfig struct {
	SnippetsFilePath string `json:"snippetsFilePath"`
	CloseToTray      bool   `json:"closeToTray"`
	TraySnippetLimit int    `json:"traySnippetLimit"`
	StartAtLogin     bool   `json:"startAtLogin"`
	Language         string `json:"language"`
}
