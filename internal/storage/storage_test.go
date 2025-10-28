package storage

import (
	"context"
	"testing"
)

func TestBuildKey(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		tag      string
		expected string
	}{
		{"No prefix", "", "v1.0.0", "v1.0.0.env"},
		{"With prefix", "envs/", "v1.0.0", "envs/v1.0.0.env"},
		{"With trailing slash", "prefix/", "test", "prefix/test.env"},
		{"No trailing slash", "prefix", "test", "prefixtest.env"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildKey(tt.prefix, tt.tag)
			if result != tt.expected {
				t.Errorf("BuildKey(%q, %q) = %q, want %q",
					tt.prefix, tt.tag, result, tt.expected)
			}
		})
	}
}

func TestMockStorageUploadDownload(t *testing.T) {
	mock := NewMockStorage()
	ctx := context.Background()

	testData := []byte("TEST_VAR=value\nANOTHER_VAR=another")
	tag := "v1.0.0"

	// Upload
	err := mock.Upload(ctx, tag, testData)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Download
	downloaded, err := mock.Download(ctx, tag)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Verify
	if string(downloaded) != string(testData) {
		t.Errorf("Downloaded data doesn't match.\nExpected: %s\nGot: %s",
			testData, downloaded)
	}
}

func TestMockStorageList(t *testing.T) {
	mock := NewMockStorage()
	ctx := context.Background()

	// Upload multiple versions
	tags := []string{"v1.0.0", "v1.1.0", "v2.0.0"}
	for _, tag := range tags {
		data := []byte("data for " + tag)
		err := mock.Upload(ctx, tag, data)
		if err != nil {
			t.Fatalf("Upload failed for %s: %v", tag, err)
		}
	}

	// List
	listed, err := mock.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Verify all tags are present
	if len(listed) != len(tags) {
		t.Errorf("Expected %d tags, got %d", len(tags), len(listed))
	}

	tagMap := make(map[string]bool)
	for _, tag := range listed {
		tagMap[tag] = true
	}

	for _, expectedTag := range tags {
		if !tagMap[expectedTag] {
			t.Errorf("Expected tag %s not found in list", expectedTag)
		}
	}
}

func TestMockStorageExists(t *testing.T) {
	mock := NewMockStorage()
	ctx := context.Background()

	tag := "v1.0.0"
	data := []byte("test data")

	// Should not exist initially
	exists, err := mock.Exists(ctx, tag)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Tag should not exist initially")
	}

	// Upload
	err = mock.Upload(ctx, tag, data)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Should exist now
	exists, err = mock.Exists(ctx, tag)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Tag should exist after upload")
	}
}

func TestMockStorageDelete(t *testing.T) {
	mock := NewMockStorage()
	ctx := context.Background()

	tag := "v1.0.0"
	data := []byte("test data")

	// Upload
	err := mock.Upload(ctx, tag, data)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Verify exists
	exists, err := mock.Exists(ctx, tag)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("Tag should exist after upload")
	}

	// Delete
	err = mock.Delete(ctx, tag)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify doesn't exist
	exists, err = mock.Exists(ctx, tag)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Tag should not exist after delete")
	}

	// Download should fail
	_, err = mock.Download(ctx, tag)
	if err == nil {
		t.Error("Expected error when downloading deleted tag")
	}
}

func TestMockStorageDownloadNonexistent(t *testing.T) {
	mock := NewMockStorage()
	ctx := context.Background()

	_, err := mock.Download(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when downloading nonexistent tag")
	}
}

func TestMockStorageError(t *testing.T) {
	mock := NewMockStorage()
	ctx := context.Background()
	expectedError := "mock error"
	mock.Error = &mockError{msg: expectedError}

	// All operations should return the error
	err := mock.Upload(ctx, "tag", []byte("data"))
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error %q, got %v", expectedError, err)
	}

	_, err = mock.Download(ctx, "tag")
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error %q, got %v", expectedError, err)
	}

	_, err = mock.List(ctx)
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error %q, got %v", expectedError, err)
	}

	_, err = mock.Exists(ctx, "tag")
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error %q, got %v", expectedError, err)
	}

	err = mock.Delete(ctx, "tag")
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error %q, got %v", expectedError, err)
	}
}

func TestMockStorageReset(t *testing.T) {
	mock := NewMockStorage()
	ctx := context.Background()

	// Upload some data
	mock.Upload(ctx, "v1.0.0", []byte("data1"))
	mock.Upload(ctx, "v2.0.0", []byte("data2"))

	// Verify data exists
	listed, _ := mock.List(ctx)
	if len(listed) != 2 {
		t.Errorf("Expected 2 items before reset, got %d", len(listed))
	}

	// Set error
	mock.Error = &mockError{msg: "test error"}

	// Reset
	mock.Reset()

	// Verify error is cleared
	if mock.Error != nil {
		t.Error("Error should be cleared after reset")
	}

	// Verify data is cleared
	listed, err := mock.List(ctx)
	if err != nil {
		t.Fatalf("List failed after reset: %v", err)
	}
	if len(listed) != 0 {
		t.Errorf("Expected no data after reset, got %d items", len(listed))
	}
}

func TestMockStorageThreadSafety(t *testing.T) {
	mock := NewMockStorage()
	ctx := context.Background()

	// Run concurrent operations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			tag := "v1.0." + string(rune('0'+n))
			data := []byte("data" + string(rune('0'+n)))

			// Upload
			mock.Upload(ctx, tag, data)

			// Download
			mock.Download(ctx, tag)

			// Exists
			mock.Exists(ctx, tag)

			// List
			mock.List(ctx)

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no panic occurred and we can still operate
	_, err := mock.List(ctx)
	if err != nil {
		t.Errorf("List failed after concurrent operations: %v", err)
	}
}

// mockError implements error interface for testing
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
