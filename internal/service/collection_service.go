package service

import (
	"errors"
	"strings"

	"SnippetsDome/internal/cryptography"
	"SnippetsDome/internal/domain"
)

func (s *SnippetService) activeCollection() (domain.Collection, error) {
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return domain.Collection{}, errors.New("collection operations are not supported")
	}
	return repository.ActiveCollection()
}

func (s *SnippetService) ActiveCollection() (domain.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return domain.Collection{}, errors.New("collection operations are not supported")
	}
	collection, err := repository.ActiveCollection()
	if err == nil && collection.Encrypted {
		collection.Locked = s.unlockedCollectionID != collection.ID
		if collection.Locked {
			collection.Name = lockedCollectionName(collection.ID)
		}
	}
	return collection, err
}

func lockedCollectionName(id string) string {
	if len(id) > 8 {
		id = id[:8]
	}
	return "Locked collection (" + id + ")"
}

func (s *SnippetService) DefaultCollection() (domain.Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return domain.Collection{}, errors.New("collection operations are not supported")
	}
	collection, err := repository.DefaultCollection()
	if err == nil && collection.Encrypted {
		collection.Locked = true
		collection.Name = lockedCollectionName(collection.ID)
	}
	return collection, err
}

func (s *SnippetService) ListCollections() ([]domain.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return nil, errors.New("collection operations are not supported")
	}
	collections, err := repository.ListCollections()
	if err == nil {
		for i := range collections {
			if collections[i].Encrypted {
				collections[i].Locked = s.unlockedCollectionID != collections[i].ID
				if collections[i].Locked {
					collections[i].Name = lockedCollectionName(collections[i].ID)
				} else {
					collections[i].Name, err = cryptography.Decrypt(s.activeKey, collections[i].Name, cryptography.CollectionAAD(collections[i].ID, collections[i].ID, "collection.name"))
					if err != nil {
						return nil, ErrCollectionAuthentication
					}
				}
			}
		}
	}
	return collections, err
}

func (s *SnippetService) CreateCollection(name string) (domain.Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearKey()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return domain.Collection{}, errors.New("collection operations are not supported")
	}
	return repository.CreateCollection(name)
}

func (s *SnippetService) RenameCollection(id, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("collection name is required")
	}
	collections, err := repository.ListCollections()
	if err != nil {
		return err
	}
	var encrypted bool
	found := false
	for _, collection := range collections {
		if collection.ID == id {
			encrypted = collection.Encrypted
			found = true
			break
		}
	}
	if !found {
		return errors.New("collection not found")
	}
	if encrypted {
		if len(s.activeKey) == 0 || s.unlockedCollectionID != id {
			return ErrCollectionLocked
		}
		sealedName, err := cryptography.Encrypt(s.activeKey, name, cryptography.CollectionAAD(id, id, "collection.name"))
		if err != nil {
			return err
		}
		return repository.RenameEncryptedCollection(id, sealedName, name)
	}
	return repository.RenameCollection(id, name)
}

func (s *SnippetService) DeleteCollection(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.unlockedCollectionID == id {
		s.clearKey()
	}
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	return repository.DeleteCollection(id)
}

func (s *SnippetService) SelectCollection(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearKey()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	return repository.SelectCollection(id)
}

func (s *SnippetService) SetDefaultCollection(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	return repository.SetDefaultCollection(id)
}
