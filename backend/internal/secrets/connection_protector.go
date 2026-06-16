package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

const SecretKeyEnv = "POSTGRESOME_SECRET_KEY"

// ConnectionProtector encrypts and decrypts stored source connection secrets.
// It derives a fixed 32-byte AES key from the configured secret string so the
// runtime can accept a human-manageable passphrase in development while still
// using authenticated encryption at rest.
type ConnectionProtector struct {
	key []byte
}

func NewConnectionProtectorFromEnv() (*ConnectionProtector, error) {
	secret := os.Getenv(SecretKeyEnv)
	if secret == "" {
		return nil, fmt.Errorf("%s environment variable is required", SecretKeyEnv)
	}

	key := sha256.Sum256([]byte(secret))
	return &ConnectionProtector{key: key[:]}, nil
}

func (p *ConnectionProtector) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return "", fmt.Errorf("failed to initialize source secret cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to initialize source secret gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate source secret nonce: %w", err)
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (p *ConnectionProtector) Decrypt(encoded string) (string, error) {
	blob, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode source secret: %w", err)
	}

	block, err := aes.NewCipher(p.key)
	if err != nil {
		return "", fmt.Errorf("failed to initialize source secret cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to initialize source secret gcm: %w", err)
	}

	if len(blob) < gcm.NonceSize() {
		return "", fmt.Errorf("stored source secret is too short")
	}

	nonce := blob[:gcm.NonceSize()]
	ciphertext := blob[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt source secret: %w", err)
	}

	return string(plaintext), nil
}
