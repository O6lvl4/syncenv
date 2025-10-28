package storage

import (
	"context"
	"fmt"
	"sync"
)

// MockStorage is a mock implementation of Storage for testing
type MockStorage struct {
	data  map[string][]byte
	mu    sync.RWMutex
	Error error // If set, all operations will return this error
}

// NewMockStorage creates a new mock storage instance
func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string][]byte),
	}
}

// Upload uploads data to mock storage
func (m *MockStorage) Upload(ctx context.Context, tag string, data []byte) error {
	if m.Error != nil {
		return m.Error
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := BuildKey("", tag)
	m.data[key] = make([]byte, len(data))
	copy(m.data[key], data)

	return nil
}

// Download downloads data from mock storage
func (m *MockStorage) Download(ctx context.Context, tag string) ([]byte, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	key := BuildKey("", tag)
	data, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("tag %s not found", tag)
	}

	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

// List returns all available tags from mock storage
func (m *MockStorage) List(ctx context.Context) ([]string, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	tags := make([]string, 0, len(m.data))
	for key := range m.data {
		// Remove .env suffix
		if len(key) > 4 && key[len(key)-4:] == ".env" {
			tag := key[:len(key)-4]
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

// Exists checks if a tag exists in mock storage
func (m *MockStorage) Exists(ctx context.Context, tag string) (bool, error) {
	if m.Error != nil {
		return false, m.Error
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	key := BuildKey("", tag)
	_, exists := m.data[key]
	return exists, nil
}

// Delete removes a tag from mock storage
func (m *MockStorage) Delete(ctx context.Context, tag string) error {
	if m.Error != nil {
		return m.Error
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := BuildKey("", tag)
	delete(m.data, key)
	return nil
}

// Reset clears all data in mock storage
func (m *MockStorage) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string][]byte)
	m.Error = nil
}
