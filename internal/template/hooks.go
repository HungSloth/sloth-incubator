package template

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunPostCreateHook executes the post-create hook if one is defined
func RunPostCreateHook(manifest *TemplateManifest, projectDir string, answers map[string]interface{}) error {
	if manifest.Hooks.PostCreate == "" {
		return nil
	}

	hookPath := filepath.Join(projectDir, manifest.Hooks.PostCreate)

	// Check if hook file exists
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		return nil // No hook file, skip silently
	}

	// Make the hook executable
	if err := os.Chmod(hookPath, 0755); err != nil {
		return fmt.Errorf("making hook executable: %w", err)
	}

	// Build environment variables from answers
	env := os.Environ()
	for key, val := range answers {
		envKey := "INCUBATOR_" + strings.ToUpper(key)
		env = append(env, fmt.Sprintf("%s=%v", envKey, val))
	}

	cmd := exec.Command(hookPath)
	cmd.Dir = projectDir
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("post-create hook failed: %w", err)
	}

	return nil
}
