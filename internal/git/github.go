package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckGHAvailable checks if gh CLI is available on PATH
func CheckGHAvailable() error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found on PATH. Please install it: https://cli.github.com")
	}
	return nil
}

// CreateRepo creates a GitHub repository using gh CLI
func CreateRepo(name string, private bool, dir string) (string, error) {
	if err := CheckGHAvailable(); err != nil {
		return "", err
	}

	visibility := "--public"
	if private {
		visibility = "--private"
	}

	cmd := exec.Command("gh", "repo", "create", name, visibility, "--source=.", "--remote=origin")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh repo create failed: %s: %w", strings.TrimSpace(string(output)), err)
	}

	// Extract repo URL from output
	repoURL := strings.TrimSpace(string(output))
	if repoURL == "" {
		repoURL = fmt.Sprintf("https://github.com/%s", name)
	}

	return repoURL, nil
}

// Push pushes the current branch to origin
func Push(dir string) error {
	cmd := exec.Command("git", "push", "-u", "origin", "HEAD")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}
