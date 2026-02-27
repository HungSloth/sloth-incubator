package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CheckGitAvailable checks if git is available on PATH
func CheckGitAvailable() error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found on PATH. Please install git: https://git-scm.com")
	}
	return nil
}

// InitRepo initializes a git repository in the given directory
func InitRepo(dir string) error {
	if err := CheckGitAvailable(); err != nil {
		return err
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init failed: %s: %w", string(output), err)
	}
	return nil
}

// InitialCommit stages all files and creates the initial commit
func InitialCommit(dir string) error {
	return CommitAll(dir, "Initial commit from sloth-incubator")
}

// HasRepo reports whether the directory already has a git repository.
func HasRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && info.IsDir()
}

// CommitAll stages all files and creates a commit with the given message.
func CommitAll(dir, message string) error {
	// git add .
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = dir
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %s: %w", string(output), err)
	}

	// git commit
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = dir
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %s: %w", string(output), err)
	}

	return nil
}
