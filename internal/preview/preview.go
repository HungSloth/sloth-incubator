package preview

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/HungSloth/sloth-incubator/internal/container"
	"gopkg.in/yaml.v3"
)

const (
	configRelPath             = ".incubator/preview/config.yaml"
	devcontainerConfigRelPath = ".devcontainer/devcontainer.json"
	previewEntrypointRelPath  = ".incubator/preview/entrypoint.sh"
)

var (
	lookPathCommand = exec.LookPath
	runCommand      = runCmd
	runOutput       = runCmdOutput
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
	applyDefaults(cfg)
	if err := validatePorts(cfg); err != nil {
		return "", err
	}

	if _, err := os.Stat(filepath.Join(projectDir, devcontainerConfigRelPath)); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("devcontainer config not found in %s", projectDir)
		}
		return "", fmt.Errorf("checking devcontainer config: %w", err)
	}

	if _, err := os.Stat(filepath.Join(projectDir, previewEntrypointRelPath)); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("preview entrypoint not found at %s", filepath.Join(projectDir, previewEntrypointRelPath))
		}
		return "", fmt.Errorf("checking preview entrypoint: %w", err)
	}

	if err := requireCommand("devcontainer"); err != nil {
		return "", err
	}

	if _, err := ensureDevcontainerUp(projectDir); err != nil {
		return "", err
	}

	if err := startPreviewProcess(projectDir, cfg); err != nil {
		return "", err
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

func ensureDevcontainerUp(projectDir string) (string, error) {
	out, err := runOutput(projectDir, "devcontainer", "up", "--workspace-folder", projectDir, "--log-format", "json")
	if err != nil {
		return "", fmt.Errorf("starting devcontainer: %w", err)
	}
	containerID := containerIDFromUpOutput(out)
	if containerID == "" {
		containerID = container.ContainerIDForProject(projectDir)
	}
	if containerID == "" {
		return "", errors.New("could not resolve devcontainer container ID after `devcontainer up`")
	}
	return containerID, nil
}

func startPreviewProcess(projectDir string, cfg *Config) error {
	cmd := fmt.Sprintf(
		"pkill -f '[x]11vnc -display' >/dev/null 2>&1 || true; "+
			"pkill -f '[w]ebsockify --web' >/dev/null 2>&1 || true; "+
			"pkill -f '[X]vfb :99' >/dev/null 2>&1 || true; "+
			"nohup env PREVIEW_APP_COMMAND=%s NOVNC_PORT=%d VNC_PORT=%d bash .incubator/preview/entrypoint.sh >/tmp/incubator-preview.log 2>&1 &",
		shellEscape(cfg.AppCommand),
		cfg.NoVNCPort,
		cfg.VNCPort,
	)

	if err := runCommand(projectDir, "devcontainer", "exec", "--workspace-folder", projectDir, "bash", "-lc", cmd); err != nil {
		return fmt.Errorf("starting preview process in devcontainer: %w", err)
	}
	return nil
}

func containerIDFromUpOutput(out string) string {
	matches := regexp.MustCompile(`"containerId"\s*:\s*"([a-f0-9]+)"`).FindAllStringSubmatch(out, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1][1]
}

func shellEscape(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func validatePorts(cfg *Config) error {
	if cfg.NoVNCPort < 1 || cfg.NoVNCPort > 65535 {
		return fmt.Errorf("invalid novnc_port: %d", cfg.NoVNCPort)
	}
	if cfg.VNCPort < 1 || cfg.VNCPort > 65535 {
		return fmt.Errorf("invalid vnc_port: %d", cfg.VNCPort)
	}
	if cfg.NoVNCPort == cfg.VNCPort {
		return errors.New("novnc_port and vnc_port must be different")
	}
	return nil
}

func requireCommand(name string) error {
	if _, err := lookPathCommand(name); err != nil {
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

func runCmdOutput(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, string(out))
	}
	return string(out), nil
}
