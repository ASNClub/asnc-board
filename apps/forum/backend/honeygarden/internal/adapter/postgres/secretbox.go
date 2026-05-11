package postgres

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const secretBoxPrefix = "enc:v1:"

type SecretBox struct {
	aead cipher.AEAD
}

func NewSecretBox(secret string) (*SecretBox, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("secretbox: secret is required")
	}
	sum := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, fmt.Errorf("secretbox: cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("secretbox: gcm: %w", err)
	}
	return &SecretBox{aead: aead}, nil
}

func (b *SecretBox) Encrypt(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("secretbox: nonce: %w", err)
	}
	ciphertext := b.aead.Seal(nonce, nonce, []byte(value), nil)
	return secretBoxPrefix + base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func (b *SecretBox) Decrypt(value string) (string, error) {
	if value == "" || !strings.HasPrefix(value, secretBoxPrefix) {
		return value, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(value, secretBoxPrefix))
	if err != nil {
		return "", fmt.Errorf("secretbox: decode: %w", err)
	}
	nonceSize := b.aead.NonceSize()
	if len(raw) < nonceSize {
		return "", errors.New("secretbox: ciphertext too short")
	}
	plaintext, err := b.aead.Open(nil, raw[:nonceSize], raw[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("secretbox: open: %w", err)
	}
	return string(plaintext), nil
}
