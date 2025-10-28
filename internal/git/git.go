package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GetCurrentTag returns the current Git tag
// It first tries to get an exact tag, then falls back to describe
func GetCurrentTag() (string, error) {
	// Try to get exact tag first
	cmd := exec.Command("git", "describe", "--tags", "--exact-match", "HEAD")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// If no exact tag, try to get the nearest tag
	cmd = exec.Command("git", "describe", "--tags", "--abbrev=0", "HEAD")
	output, err = cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	return "", fmt.Errorf("no git tag found")
}

// GetCurrentBranch returns the current Git branch name
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	if branch == "HEAD" {
		return "", fmt.Errorf("currently in detached HEAD state")
	}

	return branch, nil
}

// GetCurrentVersion returns the current version (tag or branch)
// It prefers tags over branches
func GetCurrentVersion() (string, error) {
	// Try tag first
	tag, err := GetCurrentTag()
	if err == nil && tag != "" {
		return tag, nil
	}

	// Fall back to branch
	branch, err := GetCurrentBranch()
	if err != nil {
		return "", fmt.Errorf("failed to determine current version: %w", err)
	}

	return branch, nil
}

// IsGitRepository checks if the current directory is a Git repository
func IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// GetCommitHash returns the current commit hash
func GetCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// HasUncommittedChanges checks if there are uncommitted changes
func HasUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	return len(bytes.TrimSpace(output)) > 0, nil
}

// GetAllTags returns all Git tags sorted by version
func GetAllTags() ([]string, error) {
	cmd := exec.Command("git", "tag", "-l", "--sort=-version:refname")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var tags []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			tags = append(tags, line)
		}
	}

	return tags, nil
}
