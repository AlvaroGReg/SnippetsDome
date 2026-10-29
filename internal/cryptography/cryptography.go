// Package cryptography contains the collection encryption primitives.
package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	KeySize               = 32
	SaltSize              = 16
	NonceSize             = 12
	TagSize               = 16
	FormatVersion         = "v1:"
	DefaultMemory  uint32 = 65536 // KiB
	DefaultTime    uint32 = 3
	DefaultThreads uint8  = 1
)

var (
	ErrAuthentication    = errors.New("authentication failed")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrInvalidParameters = errors.New("invalid key derivation parameters")
	ErrInvalidSalt       = errors.New("invalid salt")
	ErrInvalidKey        = errors.New("invalid encryption key")
)

// Parameters are persisted alongside the salt so a key can be derived again.
type Parameters struct {
	MemoryKiB   uint32 `json:"memoryKiB"`
	Iterations  uint32 `json:"iterations"`
	Parallelism uint8  `json:"parallelism"`
}

func DefaultParameters() Parameters {
	return Parameters{MemoryKiB: DefaultMemory, Iterations: DefaultTime, Parallelism: DefaultThreads}
}

func (p Parameters) Validate() error {
	if p.MemoryKiB < 8*uint32(p.Parallelism) || p.MemoryKiB > 262144 || p.Iterations == 0 || p.Iterations > 10 || p.Parallelism == 0 || p.Parallelism > 4 {
		return ErrInvalidParameters
	}
	return nil
}

func MarshalParameters(p Parameters) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	b, err := json.Marshal(p)
	if err != nil {
		return "", ErrInvalidParameters
	}
	return string(b), nil
}

func UnmarshalParameters(encoded string) (Parameters, error) {
	var p Parameters
	if err := json.Unmarshal([]byte(encoded), &p); err != nil {
		return Parameters{}, ErrInvalidParameters
	}
	if err := p.Validate(); err != nil {
		return Parameters{}, err
	}
	return p, nil
}

// NewSalt returns an independent cryptographically-random salt.
func NewSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		Clear(salt)
		return nil, errors.New("generate salt failed")
	}
	return salt, nil
}

// DeriveKey derives an AES-256 key from a PIN/password and validated parameters.
func DeriveKey(password string, salt []byte, p Parameters) ([]byte, error) {
	if len(salt) != SaltSize {
		return nil, ErrInvalidSalt
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return argon2.IDKey([]byte(password), salt, p.Iterations, p.MemoryKiB, p.Parallelism, KeySize), nil
}

// Clear overwrites a sensitive byte buffer as a best-effort measure.
func Clear(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
}

// CollectionAAD constructs unambiguous associated data for collection fields.
func CollectionAAD(collectionID, recordID, field string) []byte {
	return []byte(strings.Join([]string{"collection-v1", collectionID, recordID, field}, "\x00"))
}

// Encrypt seals a value and serializes nonce || ciphertext || tag as v1 Base64URL.
func Encrypt(key []byte, plaintext string, aad []byte) (string, error) {
	aead, err := newAEAD(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.New("encryption failed")
	}
	sealed := aead.Seal(nil, nonce, []byte(plaintext), aad)
	data := append(nonce, sealed...)
	return FormatVersion + base64.RawURLEncoding.EncodeToString(data), nil
}

// Decrypt authenticates and opens a serialized value. Authentication failures
// deliberately use one generic error for wrong passwords and modified data.
func Decrypt(key []byte, encoded string, aad []byte) (string, error) {
	if !strings.HasPrefix(encoded, FormatVersion) {
		return "", ErrInvalidCiphertext
	}
	data, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(encoded, FormatVersion))
	if err != nil || len(data) < NonceSize+TagSize {
		return "", ErrInvalidCiphertext
	}
	aead, err := newAEAD(key)
	if err != nil {
		return "", err
	}
	plain, err := aead.Open(nil, data[:NonceSize], data[NonceSize:], aad)
	if err != nil {
		return "", ErrAuthentication
	}
	return string(plain), nil
}

func newAEAD(key []byte) (cipher.AEAD, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w", ErrInvalidKey)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrInvalidKey
	}
	return aead, nil
}
