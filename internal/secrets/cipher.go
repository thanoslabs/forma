// Package secrets provides authenticated encryption for provider credentials.
package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

type Cipher struct{ aead cipher.AEAD }

func New(key []byte) (*Cipher, error) {
	if len(key) != 32 {
		return nil, errors.New("credential encryption key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{aead: aead}, nil
}

func FromBase64(value string) (*Cipher, error) {
	key, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode credential encryption key: %w", err)
	}
	return New(key)
}

func (c *Cipher) Seal(plaintext, associatedData []byte) ([]byte, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return c.aead.Seal(nonce, nonce, plaintext, associatedData), nil
}

func (c *Cipher) Open(ciphertext, associatedData []byte) ([]byte, error) {
	if len(ciphertext) < c.aead.NonceSize() {
		return nil, errors.New("credential ciphertext is invalid")
	}
	nonce := ciphertext[:c.aead.NonceSize()]
	return c.aead.Open(nil, nonce, ciphertext[c.aead.NonceSize():], associatedData)
}
