package service

import (
	"errors"
	"unicode/utf8"

	"SnippetsDome/internal/cryptography"
	"SnippetsDome/internal/domain"
)

func validCollectionPassword(password string) bool {
	length := utf8.RuneCountInString(password)
	return utf8.ValidString(password) && length >= 1 && length <= 10
}

func (s *SnippetService) requireUnlockedActiveCollection() error {
	collection, err := s.activeCollection()
	if err != nil {
		return err
	}
	if collection.Encrypted && (len(s.activeKey) == 0 || s.unlockedCollectionID != collection.ID) {
		return ErrCollectionLocked
	}
	return nil
}

func (s *SnippetService) EncryptCollection(id, password, confirmation string) error {
	if !validCollectionPassword(password) || !validCollectionPassword(confirmation) || password != confirmation {
		return ErrInvalidCollectionPassword
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	salt, err := cryptography.NewSalt()
	if err != nil {
		return err
	}
	parameters := cryptography.DefaultParameters()
	key, err := cryptography.DeriveKey(password, salt, parameters)
	if err != nil {
		return err
	}
	defer cryptography.Clear(key)
	check, err := cryptography.Encrypt(key, "collection-key-check-v1", cryptography.CollectionAAD(id, id, "collection.key-check"))
	if err != nil {
		return err
	}
	if err := repository.EncryptCollection(id, key, salt, parameters, check); err != nil {
		return err
	}
	collection, err := repository.ActiveCollection()
	if err == nil && collection.ID == id {
		s.clearKey()
		s.activeKey = append([]byte(nil), key...)
		s.unlockedCollectionID = id
	}
	return nil
}

func (s *SnippetService) UnlockCollection(id, password string) (domain.Collection, error) {
	if !validCollectionPassword(password) {
		return domain.Collection{}, ErrInvalidCollectionPassword
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return domain.Collection{}, errors.New("collection operations are not supported")
	}
	s.clearKey()
	if err := repository.SelectCollection(id); err != nil {
		return domain.Collection{}, err
	}
	salt, encodedParameters, check, err := repository.CollectionEncryptionMetadata(id)
	if err != nil {
		return domain.Collection{}, err
	}
	parameters, err := cryptography.UnmarshalParameters(encodedParameters)
	if err != nil {
		return domain.Collection{}, ErrCollectionAuthentication
	}
	key, err := cryptography.DeriveKey(password, salt, parameters)
	if err != nil {
		return domain.Collection{}, ErrCollectionAuthentication
	}
	verified, err := cryptography.Decrypt(key, check, cryptography.CollectionAAD(id, id, "collection.key-check"))
	if err != nil || verified != "collection-key-check-v1" {
		cryptography.Clear(key)
		return domain.Collection{}, ErrCollectionAuthentication
	}
	s.activeKey = key
	s.unlockedCollectionID = id
	collection, err := repository.ActiveCollection()
	if err != nil {
		s.clearKey()
		return domain.Collection{}, err
	}
	collection.Name, err = cryptography.Decrypt(key, collection.Name, cryptography.CollectionAAD(id, id, "collection.name"))
	if err != nil {
		s.clearKey()
		return domain.Collection{}, ErrCollectionAuthentication
	}
	return collection, nil
}

func (s *SnippetService) clearKey() {
	cryptography.Clear(s.activeKey)
	s.activeKey = nil
	s.unlockedCollectionID = ""
}

func (s *SnippetService) LockActiveCollection() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearKey()
	return nil
}

func (s *SnippetService) DecryptCollection(id, password string) error {
	if !validCollectionPassword(password) {
		return ErrInvalidCollectionPassword
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	repository, ok := s.repository.(collectionRepository)
	if !ok {
		return errors.New("collection operations are not supported")
	}
	salt, encodedParameters, check, err := repository.CollectionEncryptionMetadata(id)
	if err != nil {
		return err
	}
	parameters, err := cryptography.UnmarshalParameters(encodedParameters)
	if err != nil {
		return ErrCollectionAuthentication
	}
	key, err := cryptography.DeriveKey(password, salt, parameters)
	if err != nil {
		return ErrCollectionAuthentication
	}
	defer cryptography.Clear(key)
	verified, err := cryptography.Decrypt(key, check, cryptography.CollectionAAD(id, id, "collection.key-check"))
	if err != nil || verified != "collection-key-check-v1" {
		return ErrCollectionAuthentication
	}
	if err := repository.DecryptCollection(id, key); err != nil {
		return err
	}
	if s.unlockedCollectionID == id {
		s.clearKey()
	}
	return nil
}

func transformSnippet(collectionID string, key []byte, snippet *domain.Snippet, encrypt bool) error {
	fields := []struct {
		value *string
		name  string
	}{{&snippet.Title, "snippet.title"}, {&snippet.Language, "snippet.language"}, {&snippet.Code, "snippet.code"}}
	for _, field := range fields {
		var result string
		var err error
		aad := cryptography.CollectionAAD(collectionID, snippet.ID, field.name)
		if encrypt {
			result, err = cryptography.Encrypt(key, *field.value, aad)
		} else {
			result, err = cryptography.Decrypt(key, *field.value, aad)
		}
		if err != nil {
			return err
		}
		*field.value = result
	}
	for i, tag := range snippet.Tags {
		var err error
		aad := cryptography.CollectionAAD(collectionID, snippet.ID, "snippet.tag")
		if encrypt {
			snippet.Tags[i], err = cryptography.Encrypt(key, tag, aad)
		} else {
			snippet.Tags[i], err = cryptography.Decrypt(key, tag, aad)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
