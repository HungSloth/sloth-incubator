package preview

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"

	"gopkg.in/yaml.v3"
)

const (
	configRelPath     = ".incubator/preview/config.yaml"
	dockerfileRelPath = ".incubator/preview/Dockerfile"
)

// Config controls the local noVNC preview runtime.
type Config struct {
	Enabled    bool   `yaml:"enabled"`
	AppCommand string `yaml:"app_command"`
	NoVNCPort  int    `yaml:"novnc_port"`
	VNCPort    int    `yaml:"vnc_port"`
}

// LoadConfig loads preview configuration from a generated project.
func LoadConfig(projectDir string) (*Config, error) {
	configPath := filepath.Join(projectDir, configRelPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("preview config not found at %s", configPath)
		}
		return nil, fmt.Errorf("reading preview config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing preview config: %w", err)
	}
	applyDefaults(&cfg)
	return &cfg, nil
}

// Start starts or replaces a preview container and returns the noVNC URL.
func Start(projectDir string, cfg *Config) (string, error) {
	if cfg == nil {
		return "", errors.New("preview config is required")
	}
	if !cfg.Enabled {
		return "", errors.New("preview is disabled in config")
	}

	if _, err := os.Stat(filepath.Join(projectDir, dockerfileRelPath)); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("preview Dockerfile not found in %s", projectDir)
		}
		return "", fmt.Errorf("checking preview Dockerfile: %w", err)
	}

	if err := requireCommand("docker"); err != nil {
		return "", err
	}

	containerName := containerNameForProject(projectDir)
	imageName := containerName + "-image"
	if err := runCmd(projectDir, "docker", "build", "-t", imageName, "-f", dockerfileRelPath, "."); err != nil {
		return "", fmt.Errorf("building preview image: %w", err)
	}

	// Remove an old container with the same name, if any.
	_ = runCmd(projectDir, "docker", "rm", "-f", containerName)

	if err := runCmd(projectDir, "docker",
		"run", "-d",
		"--name", containerName,
		"-p", fmt.Sprintf("%d:%d", cfg.NoVNCPort, cfg.NoVNCPort),
		"-p", fmt.Sprintf("%d:%d", cfg.VNCPort, cfg.VNCPort),
		"-e", fmt.Sprintf("PREVIEW_APP_COMMAND=%s", cfg.AppCommand),
		"-e", fmt.Sprintf("NOVNC_PORT=%d", cfg.NoVNCPort),
		"-e", fmt.Sprintf("VNC_PORT=%d", cfg.VNCPort),
		"-v", fmt.Sprintf("%s:/workspace", projectDir),
		imageName,
	); err != nil {
		return "", fmt.Errorf("starting preview container: %w", err)
	}

	return fmt.Sprintf("http://localhost:%d", cfg.NoVNCPort), nil
}

// OpenBrowser opens the provided URL in the default local browser.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func applyDefaults(cfg *Config) {
	if cfg.NoVNCPort == 0 {
		cfg.NoVNCPort = 6080
	}
	if cfg.VNCPort == 0 {
		cfg.VNCPort = 5900
	}
}

func containerNameForProject(projectDir string) string {
	base := filepath.Base(projectDir)
	safe := regexp.MustCompile(`[^a-zA-Z0-9_.-]+`).ReplaceAllString(base, "-")
	if safe == "" {
		safe = "project"
	}
	return "sloth-preview-" + safe
}

func requireCommand(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("required command not found: %s", name)
	}
	return nil
}

func runCmd(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
