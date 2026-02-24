package preview

import (
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

func TestStartPreflightRequiresDockerfile(t *testing.T) {
	projectDir := t.TempDir()
	_, err := Start(projectDir, &Config{Enabled: true, NoVNCPort: 6080, VNCPort: 5900})
	if err == nil {
		t.Fatal("expected error when Dockerfile is missing")
	}
	if !strings.Contains(err.Error(), "preview Dockerfile not found") {
		t.Fatalf("expected missing Dockerfile error, got: %v", err)
	}
}
