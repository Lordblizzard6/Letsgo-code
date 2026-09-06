package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

const encPrefix = "enc:v1:"

// getEncryptionKey derives a machine/user specific 32-byte key for AES-256.
func getEncryptionKey() []byte {
	hostname, _ := os.Hostname()
	homeDir, _ := os.UserHomeDir()
	seed := fmt.Sprintf("letsgo-code-salt:%s:%s", hostname, homeDir)
	hash := sha256.Sum256([]byte(seed))
	return hash[:]
}

// EncryptSecret encrypts a plaintext string using AES-256-GCM.
// Returns encPrefix + base64(nonce + ciphertext).
func EncryptSecret(plaintext string) (string, error) {
	if plaintext == "" || strings.HasPrefix(plaintext, encPrefix) {
		return plaintext, nil
	}
	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return encPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptSecret decrypts a string if it starts with encPrefix.
// If it does not start with encPrefix, returns it as-is (plaintext fallback).
func DecryptSecret(encrypted string) (string, error) {
	if !strings.HasPrefix(encrypted, encPrefix) {
		return encrypted, nil
	}
	raw := strings.TrimPrefix(encrypted, encPrefix)
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}
	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}
	return string(plaintext), nil
}
