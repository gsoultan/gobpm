package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptionKey is used for encrypting sensitive process variables.
// In production, this MUST be loaded from a secure environment variable or KMS.
var EncryptionKey = []byte("gobpm-engine-32bytes-secret-key-") // 32 bytes for AES-256

// DeriveKey derives a 32-byte AES-256 key from an arbitrary-length passphrase using SHA-256.
// Always call this with a secret loaded from a secure source (environment variable, KMS, etc.).
func DeriveKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:]
}

// Encrypt encrypts plaintext using the global EncryptionKey (AES-256-GCM).
func Encrypt(plaintext string) (string, error) {
	return EncryptWithKey(plaintext, EncryptionKey)
}

// Decrypt decrypts ciphertext using the global EncryptionKey (AES-256-GCM).
func Decrypt(ciphertextStr string) (string, error) {
	return DecryptWithKey(ciphertextStr, EncryptionKey)
}

// EncryptWithKey encrypts plaintext using the provided key (AES-256-GCM).
// The key must be exactly 32 bytes for AES-256.
func EncryptWithKey(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptWithKey decrypts ciphertext using the provided key (AES-256-GCM).
// The key must be exactly 32 bytes for AES-256.
func DecryptWithKey(ciphertextStr string, key []byte) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
