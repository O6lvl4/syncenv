package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/O6lvl4/syncenv/internal/archive"
	"github.com/O6lvl4/syncenv/internal/config"
	"github.com/O6lvl4/syncenv/internal/crypto"
)

// loadEnvFiles reads multiple env files and returns them as an archive
func loadEnvFiles(cfg *config.Config) ([]byte, error) {
	files := cfg.GetEnvFiles()

	// Check if all files exist
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", file)
		}
	}

	// If only one file, just read it directly (for backward compatibility)
	if len(files) == 1 {
		data, err := os.ReadFile(files[0])
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", files[0], err)
		}
		return data, nil
	}

	// Multiple files: create archive
	archiveData, err := archive.Create(files)
	if err != nil {
		return nil, fmt.Errorf("failed to create archive: %w", err)
	}

	return archiveData, nil
}

// prepareData prepares data for upload (encrypts if needed)
func prepareData(data []byte, cfg *config.Config) ([]byte, error) {
	if !cfg.Encryption.Enabled {
		return data, nil
	}

	if cfg.Encryption.Key == "" {
		return nil, fmt.Errorf("encryption is enabled but no key is configured")
	}

	key, err := crypto.DecodeKeyFromString(cfg.Encryption.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	encrypted, err := crypto.Encrypt(data, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	return encrypted, nil
}

// processData processes downloaded data (decrypts if needed)
func processData(data []byte, cfg *config.Config) ([]byte, error) {
	if !cfg.Encryption.Enabled || cfg.Encryption.Key == "" {
		return data, nil
	}

	key, err := crypto.DecodeKeyFromString(cfg.Encryption.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	decrypted, err := crypto.Decrypt(data, key)
	if err != nil {
		// If decryption fails, the data might not be encrypted
		// Return the original data as-is
		return data, nil
	}

	return decrypted, nil
}

// saveEnvFiles writes data to env files (extracts archive if multiple files)
func saveEnvFiles(data []byte, cfg *config.Config) error {
	files := cfg.GetEnvFiles()

	// If only one file, just write it directly (for backward compatibility)
	if len(files) == 1 {
		if err := os.WriteFile(files[0], data, 0600); err != nil {
			return fmt.Errorf("failed to write file %s: %w", files[0], err)
		}
		return nil
	}

	// Multiple files: extract archive
	if err := archive.ExtractToFiles(data); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return nil
}

// parseEnvFile parses env file content into a map
func parseEnvFile(data []byte) (map[string]string, error) {
	envMap := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			envMap[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse env file: %w", err)
	}

	return envMap, nil
}

// parseDataToEnvMap parses data (single file or archive) into an env map
func parseDataToEnvMap(data []byte, cfg *config.Config) (map[string]string, error) {
	files := cfg.GetEnvFiles()

	// If single file, parse directly
	if len(files) == 1 {
		return parseEnvFile(data)
	}

	// Multiple files: extract archive and parse all .env files
	entries, err := archive.Extract(data)
	if err != nil {
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	envMap := make(map[string]string)
	for _, entry := range entries {
		// Only parse .env files, skip JSON and other formats
		if !strings.HasSuffix(entry.Path, ".env") && !strings.HasPrefix(entry.Path, ".env") {
			continue
		}

		fileEnv, err := parseEnvFile(entry.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Path, err)
		}

		// Merge into main map
		for key, value := range fileEnv {
			envMap[key] = value
		}
	}

	return envMap, nil
}

// formatEnvFile formats env map back to file content
func formatEnvFile(envMap map[string]string) string {
	var lines []string
	for key, value := range envMap {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(lines, "\n") + "\n"
}

// diffEnvMaps compares two env maps and returns added, removed, and changed keys
func diffEnvMaps(oldMap, newMap map[string]string) (added, removed, changed map[string]string) {
	added = make(map[string]string)
	removed = make(map[string]string)
	changed = make(map[string]string)

	// Find added and changed
	for key, newValue := range newMap {
		if oldValue, exists := oldMap[key]; exists {
			if oldValue != newValue {
				changed[key] = fmt.Sprintf("%s -> %s", oldValue, newValue)
			}
		} else {
			added[key] = newValue
		}
	}

	// Find removed
	for key, value := range oldMap {
		if _, exists := newMap[key]; !exists {
			removed[key] = value
		}
	}

	return added, removed, changed
}
