package cryptography

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	p := DefaultParameters()
	salt, err := NewSalt()
	if err != nil {
		t.Fatal(err)
	}
	key, err := DeriveKey("0042", salt, p)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { Clear(key) })

	for _, value := range []string{"", "short", strings.Repeat("long-value-", 10000)} {
		encoded, err := Encrypt(key, value, CollectionAAD("collection", "snippet", "code"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(encoded, FormatVersion) {
			t.Fatalf("ciphertext = %q", encoded)
		}
		got, err := Decrypt(key, encoded, CollectionAAD("collection", "snippet", "code"))
		if err != nil || got != value {
			t.Fatalf("round trip got %q, error = %v", got, err)
		}
	}
}

func TestDecryptRejectsWrongAADAndModifiedCiphertext(t *testing.T) {
	key := make([]byte, KeySize)
	encoded, err := Encrypt(key, "secret", []byte("right-aad"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Decrypt(key, encoded, []byte("wrong-aad")); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("wrong AAD error = %v", err)
	}
	wrongKey := make([]byte, KeySize)
	wrongKey[0] = 1
	if _, err := Decrypt(wrongKey, encoded, []byte("right-aad")); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("wrong key error = %v", err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(encoded, FormatVersion))
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-1] ^= 1
	altered := FormatVersion + base64.RawURLEncoding.EncodeToString(raw)
	if _, err := Decrypt(key, altered, []byte("right-aad")); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("altered ciphertext error = %v", err)
	}
}

func TestDecryptRejectsMalformedCiphertext(t *testing.T) {
	key := make([]byte, KeySize)
	for _, value := range []string{"", "v2:abc", "v1:!", "v1:" + base64.RawURLEncoding.EncodeToString(make([]byte, NonceSize+TagSize-1))} {
		if _, err := Decrypt(key, value, nil); !errors.Is(err, ErrInvalidCiphertext) {
			t.Errorf("Decrypt(%q) error = %v", value, err)
		}
	}
}

func TestParametersSerializationAndValidation(t *testing.T) {
	want := DefaultParameters()
	encoded, err := MarshalParameters(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalParameters(encoded)
	if err != nil || got != want {
		t.Fatalf("parameters = %#v, error = %v", got, err)
	}
	for _, p := range []Parameters{{}, {MemoryKiB: 262145, Iterations: 1, Parallelism: 1}, {MemoryKiB: 8, Iterations: 11, Parallelism: 1}, {MemoryKiB: 8, Iterations: 1, Parallelism: 5}} {
		if err := p.Validate(); !errors.Is(err, ErrInvalidParameters) {
			t.Errorf("Validate(%#v) = %v", p, err)
		}
	}
	if _, err := UnmarshalParameters(`{"memoryKiB":4294967295,"iterations":3,"parallelism":1}`); !errors.Is(err, ErrInvalidParameters) {
		t.Fatalf("invalid serialized parameters error = %v", err)
	}
}

func TestDeriveKeyRejectsInvalidSaltAndParameters(t *testing.T) {
	if _, err := DeriveKey("1234", []byte("short"), DefaultParameters()); !errors.Is(err, ErrInvalidSalt) {
		t.Fatalf("salt error = %v", err)
	}
	if _, err := DeriveKey("1234", make([]byte, SaltSize), Parameters{}); !errors.Is(err, ErrInvalidParameters) {
		t.Fatalf("parameters error = %v", err)
	}
}

func TestClearOverwritesBuffer(t *testing.T) {
	buf := []byte("secret")
	Clear(buf)
	for i, value := range buf {
		if value != 0 {
			t.Fatalf("byte %d was not cleared", i)
		}
	}
}
