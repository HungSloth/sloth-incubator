package preview

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigAppliesPortDefaults(t *testing.T) {
	projectDir := t.TempDir()
	configDir := filepath.Join(projectDir, ".incubator", "preview")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("creating config dir: %v", err)
	}

	configContent := "enabled: true\napp_command: \"echo hi\"\n"
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	cfg, err := LoadConfig(projectDir)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.NoVNCPort != 6080 {
		t.Fatalf("expected default novnc port 6080, got %d", cfg.NoVNCPort)
	}
	if cfg.VNCPort != 5900 {
		t.Fatalf("expected default vnc port 5900, got %d", cfg.VNCPort)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	projectDir := t.TempDir()
	_, err := LoadConfig(projectDir)
	if err == nil {
		t.Fatal("expected error when config file is missing")
	}
	if !strings.Contains(err.Error(), "preview config not found") {
		t.Fatalf("expected missing config error, got: %v", err)
	}
}

func TestStartReturnsErrorForDisabledPreview(t *testing.T) {
	projectDir := t.TempDir()
	_, err := Start(projectDir, &Config{Enabled: false})
	if err == nil {
		t.Fatal("expected error when preview is disabled")
	}
}

func TestStartPreflightRequiresDevcontainer(t *testing.T) {
	projectDir := t.TempDir()
	_, err := Start(projectDir, &Config{Enabled: true, NoVNCPort: 6080, VNCPort: 5900})
	if err == nil {
		t.Fatal("expected error when devcontainer is missing")
	}
	if !strings.Contains(err.Error(), "devcontainer config not found") {
		t.Fatalf("expected missing devcontainer error, got: %v", err)
	}
}

func TestStartReturnsErrorWhenDevcontainerCLIIsMissing(t *testing.T) {
	projectDir := writePreviewProject(t)

	origLookPath := lookPathCommand
	defer func() { lookPathCommand = origLookPath }()
	lookPathCommand = func(name string) (string, error) {
		return "", fmt.Errorf("not found")
	}

	_, err := Start(projectDir, &Config{Enabled: true, NoVNCPort: 6080, VNCPort: 5900})
	if err == nil {
		t.Fatal("expected error when devcontainer CLI is missing")
	}
	if !strings.Contains(err.Error(), "required command not found: devcontainer") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStartUsesDevcontainerRuntime(t *testing.T) {
	projectDir := writePreviewProject(t)

	origLookPath := lookPathCommand
	origRunOutput := runOutput
	origRunCommand := runCommand
	defer func() {
		lookPathCommand = origLookPath
		runOutput = origRunOutput
		runCommand = origRunCommand
	}()

	lookPathCommand = func(name string) (string, error) { return "/usr/bin/" + name, nil }

	upCalled := false
	execCalled := false

	runOutput = func(dir, name string, args ...string) (string, error) {
		if name == "devcontainer" && len(args) >= 1 && args[0] == "up" {
			upCalled = true
			return `{"containerId":"abc123def456"}`, nil
		}
		return "", nil
	}
	runCommand = func(dir, name string, args ...string) error {
		if name == "devcontainer" && len(args) >= 1 && args[0] == "exec" {
			execCalled = true
			if !strings.Contains(strings.Join(args, " "), ".incubator/preview/entrypoint.sh") {
				t.Fatalf("expected preview entrypoint in exec command, got: %v", args)
			}
		}
		return nil
	}

	url, err := Start(projectDir, &Config{
		Enabled:    true,
		AppCommand: "npm run dev",
		NoVNCPort:  6080,
		VNCPort:    5900,
	})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if url != "http://localhost:6080" {
		t.Fatalf("unexpected URL %q", url)
	}
	if !upCalled {
		t.Fatal("expected devcontainer up to be called")
	}
	if !execCalled {
		t.Fatal("expected devcontainer exec to be called")
	}
}

func writePreviewProject(t *testing.T) string {
	t.Helper()
	projectDir := t.TempDir()
	devcontainerDir := filepath.Join(projectDir, ".devcontainer")
	previewDir := filepath.Join(projectDir, ".incubator", "preview")
	if err := os.MkdirAll(devcontainerDir, 0755); err != nil {
		t.Fatalf("creating devcontainer dir: %v", err)
	}
	if err := os.MkdirAll(previewDir, 0755); err != nil {
		t.Fatalf("creating preview dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(devcontainerDir, "devcontainer.json"), []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatalf("writing devcontainer config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(previewDir, "entrypoint.sh"), []byte("#!/usr/bin/env bash\necho test\n"), 0755); err != nil {
		t.Fatalf("writing preview entrypoint: %v", err)
	}
	return projectDir
}
