package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileEntry represents a file in the archive
type FileEntry struct {
	Path string
	Data []byte
	Mode os.FileMode
}

// Create creates a tar.gz archive from multiple files
func Create(files []string) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)

	for _, file := range files {
		// Read file
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		// Get file info
		info, err := os.Stat(file)
		if err != nil {
			return nil, fmt.Errorf("failed to stat file %s: %w", file, err)
		}

		// Create tar header
		header := &tar.Header{
			Name: file,
			Mode: int64(info.Mode()),
			Size: int64(len(data)),
		}

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("failed to write tar header for %s: %w", file, err)
		}

		// Write file data
		if _, err := tarWriter.Write(data); err != nil {
			return nil, fmt.Errorf("failed to write file data for %s: %w", file, err)
		}
	}

	// Close writers
	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Extract extracts a tar.gz archive to multiple files
func Extract(archiveData []byte) ([]FileEntry, error) {
	var entries []FileEntry

	// Create gzip reader
	gzReader, err := gzip.NewReader(bytes.NewReader(archiveData))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Read all entries
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Read file data
		data := make([]byte, header.Size)
		if _, err := io.ReadFull(tarReader, data); err != nil {
			return nil, fmt.Errorf("failed to read file data for %s: %w", header.Name, err)
		}

		entries = append(entries, FileEntry{
			Path: header.Name,
			Data: data,
			Mode: os.FileMode(header.Mode),
		})
	}

	return entries, nil
}

// ExtractToFiles extracts archive and writes files to disk
func ExtractToFiles(archiveData []byte) error {
	entries, err := Extract(archiveData)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Create directory if needed
		dir := filepath.Dir(entry.Path)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}

		// Write file
		if err := os.WriteFile(entry.Path, entry.Data, entry.Mode); err != nil {
			return fmt.Errorf("failed to write file %s: %w", entry.Path, err)
		}
	}

	return nil
}

// ListFiles returns the list of files in an archive
func ListFiles(archiveData []byte) ([]string, error) {
	entries, err := Extract(archiveData)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		files = append(files, entry.Path)
	}

	return files, nil
}
