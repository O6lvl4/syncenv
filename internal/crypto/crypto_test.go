package crypto

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	if len(key) != KeySize {
		t.Errorf("Expected key size %d, got %d", KeySize, len(key))
	}
}

func TestSaveAndLoadKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")

	// Generate and save key
	originalKey, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	err = SaveKey(keyPath, originalKey)
	if err != nil {
		t.Fatalf("SaveKey failed: %v", err)
	}

	// Load key
	loadedKey, err := LoadKey(keyPath)
	if err != nil {
		t.Fatalf("LoadKey failed: %v", err)
	}

	// Compare keys
	if !bytes.Equal(originalKey, loadedKey) {
		t.Error("Loaded key doesn't match original key")
	}
}

func TestLoadKeyInvalidFile(t *testing.T) {
	_, err := LoadKey("/nonexistent/path/test.key")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestLoadKeyInvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "invalid.key")

	// Write invalid data
	err := os.WriteFile(keyPath, []byte("not-hex-data!"), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = LoadKey(keyPath)
	if err == nil {
		t.Error("Expected error for invalid key format, got nil")
	}
}

func TestLoadKeyInvalidSize(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "short.key")

	// Write short key (16 bytes hex encoded = 32 chars)
	err := os.WriteFile(keyPath, []byte("0123456789abcdef0123456789abcdef"), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = LoadKey(keyPath)
	if err == nil {
		t.Error("Expected error for invalid key size, got nil")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{"Empty data", []byte{}},
		{"Simple text", []byte("Hello, World!")},
		{"Multi-line", []byte("Line 1\nLine 2\nLine 3")},
		{"Binary data", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}},
		{"Large data", make([]byte, 10000)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := Encrypt(tc.plaintext, key)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// Ciphertext should be different from plaintext
			if len(tc.plaintext) > 0 && bytes.Equal(ciphertext, tc.plaintext) {
				t.Error("Ciphertext is identical to plaintext")
			}

			// Decrypt
			decrypted, err := Decrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			// Decrypted should match original
			if !bytes.Equal(decrypted, tc.plaintext) {
				t.Errorf("Decrypted data doesn't match original.\nExpected: %v\nGot: %v",
					tc.plaintext, decrypted)
			}
		})
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	plaintext := []byte("Secret message")

	// Encrypt with key1
	ciphertext, err := Encrypt(plaintext, key1)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Try to decrypt with key2
	_, err = Decrypt(ciphertext, key2)
	if err == nil {
		t.Error("Expected error when decrypting with wrong key, got nil")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	key, _ := GenerateKey()

	testCases := []struct {
		name       string
		ciphertext []byte
	}{
		{"Empty", []byte{}},
		{"Too short", []byte{0x01, 0x02}},
		{"Invalid data", []byte("not encrypted data")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Decrypt(tc.ciphertext, key)
			if err == nil {
				t.Error("Expected error for invalid ciphertext, got nil")
			}
		})
	}
}

func TestEncryptFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("Test file content\nLine 2\nLine 3")

	// Create test file
	err := os.WriteFile(testFile, testContent, 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	// Encrypt file
	ciphertext, err := EncryptFile(testFile, key)
	if err != nil {
		t.Fatalf("EncryptFile failed: %v", err)
	}

	// Decrypt to verify
	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, testContent) {
		t.Error("Decrypted content doesn't match original file content")
	}
}

func TestDecryptToFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.txt")
	testContent := []byte("Test content to decrypt")

	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	// Encrypt
	ciphertext, err := Encrypt(testContent, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Decrypt to file
	err = DecryptToFile(ciphertext, key, outputFile)
	if err != nil {
		t.Fatalf("DecryptToFile failed: %v", err)
	}

	// Read file and verify
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !bytes.Equal(content, testContent) {
		t.Error("File content doesn't match original")
	}
}

func TestEncryptDecryptDeterministic(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := []byte("Same plaintext")

	// Encrypt twice
	ciphertext1, _ := Encrypt(plaintext, key)
	ciphertext2, _ := Encrypt(plaintext, key)

	// Ciphertexts should be different (due to random nonce)
	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("Two encryptions of same plaintext produced identical ciphertext (nonce not random)")
	}

	// Both should decrypt to same plaintext
	decrypted1, _ := Decrypt(ciphertext1, key)
	decrypted2, _ := Decrypt(ciphertext2, key)

	if !bytes.Equal(decrypted1, plaintext) || !bytes.Equal(decrypted2, plaintext) {
		t.Error("Decryption failed for deterministic test")
	}
}
