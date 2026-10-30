package service

import (
	"errors"
	"sync"

	"SnippetsDome/internal/cryptography"
	"SnippetsDome/internal/domain"
)

var (
	ErrInvalidCollectionPassword = errors.New("collection password must contain between 1 and 10 characters")
	ErrCollectionLocked          = errors.New("collection is locked")
	ErrCollectionAuthentication  = errors.New("authentication failed")
)

// SnippetService coordinates snippet, collection, encryption, and preference
// operations. Its methods are split across focused files in this package.
type SnippetService struct {
	mu                   sync.RWMutex
	repository           snippetRepository
	config               domain.AppConfig
	configRepository     configRepository
	activeKey            []byte
	unlockedCollectionID string
}

type configRepository interface {
	SaveConfig(domain.AppConfig) error
}

type snippetRepository interface {
	ListSnippets() ([]domain.Snippet, error)
	CreateSnippet(domain.Snippet) (domain.Snippet, error)
	UpdateSnippet(domain.Snippet) (domain.Snippet, error)
	DeleteSnippet(string) error
}

type collectionRepository interface {
	ActiveCollection() (domain.Collection, error)
	DefaultCollection() (domain.Collection, error)
	ListCollections() ([]domain.Collection, error)
	CreateCollection(string) (domain.Collection, error)
	RenameCollection(string, string) error
	DeleteCollection(string) error
	SelectCollection(string) error
	SetDefaultCollection(string) error
	RenameEncryptedCollection(string, string, string) error
	EncryptCollection(string, []byte, []byte, cryptography.Parameters, string) error
	DecryptCollection(string, []byte) error
	CollectionEncryptionMetadata(string) ([]byte, string, string, error)
}

func NewSnippetService(config domain.AppConfig, snippetRepository snippetRepository, configRepository configRepository) *SnippetService {
	if config.TraySnippetLimit < 1 {
		config.TraySnippetLimit = domain.DefaultTraySnippetLimit
	}
	if config.Language == "system" {
		// Migrate the preference used by the previous system-language option.
		config.Language = ""
	}
	if !validLanguage(config.Language) && config.Language != "" {
		config.Language = domain.LanguageEnglish
	}

	return &SnippetService{
		repository:       snippetRepository,
		config:           config,
		configRepository: configRepository,
	}
}
