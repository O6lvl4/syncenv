package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const (
	KeySize = 32 // AES-256
)

// GenerateKey generates a new random encryption key
func GenerateKey() ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

// EncodeKeyToString encodes a key to a hex string for storage
func EncodeKeyToString(key []byte) string {
	return hex.EncodeToString(key)
}

// DecodeKeyFromString decodes a hex string to a key
func DecodeKeyFromString(keyHex string) ([]byte, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d bytes", KeySize, len(key))
	}

	return key, nil
}

// SaveKey saves the encryption key to a file
func SaveKey(keyPath string, key []byte) error {
	keyHex := hex.EncodeToString(key)
	if err := os.WriteFile(keyPath, []byte(keyHex), 0600); err != nil {
		return fmt.Errorf("failed to save key: %w", err)
	}
	return nil
}

// LoadKey loads the encryption key from a file
func LoadKey(keyPath string) ([]byte, error) {
	keyHex, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	key, err := hex.DecodeString(string(keyHex))
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", KeySize, len(key))
	}

	return key, nil
}

// Encrypt encrypts data using AES-256-GCM
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-256-GCM
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// EncryptFile encrypts a file and returns the encrypted data
func EncryptFile(filePath string, key []byte) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return Encrypt(data, key)
}

// DecryptToFile decrypts data and writes it to a file
func DecryptToFile(ciphertext []byte, key []byte, filePath string) error {
	plaintext, err := Decrypt(ciphertext, key)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, plaintext, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
