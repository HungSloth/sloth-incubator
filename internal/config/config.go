package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the user configuration
type Config struct {
	GitHubUser        string   `yaml:"github_user"`
	DefaultVisibility string   `yaml:"default_visibility"`
	DefaultLicense    string   `yaml:"default_license"`
	ProjectDir        string   `yaml:"project_dir"`
	Editor            string   `yaml:"editor"`
	TemplateRepo      string   `yaml:"template_repo"`
	TemplateRepos     []string `yaml:"template_repos,omitempty"`
	AutoUpdateCheck   bool     `yaml:"auto_update_check"`
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		DefaultVisibility: "private",
		DefaultLicense:    "MIT",
		ProjectDir:        filepath.Join(homeDir, "projects"),
		Editor:            "none",
		TemplateRepo:      "HungSloth/incubator-templates",
		AutoUpdateCheck:   true,
	}
}

// ConfigDir returns the path to the config directory
func ConfigDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".incubator")
}

// ConfigPath returns the path to the config file
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// Load loads the config from disk, creating defaults if it doesn't exist
func Load() (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			// Try to detect GitHub user
			cfg.GitHubUser = detectGitHubUser()
			// Save the defaults
			if saveErr := cfg.Save(); saveErr != nil {
				return cfg, nil // return defaults even if save fails
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

// Save writes the config to disk
func (c *Config) Save() error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(ConfigPath(), data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// GetProjectDir returns the expanded project directory
func (c *Config) GetProjectDir() string {
	dir := c.ProjectDir
	if strings.HasPrefix(dir, "~") {
		homeDir, _ := os.UserHomeDir()
		dir = filepath.Join(homeDir, dir[1:])
	}
	return dir
}

// GetTemplateRepos returns all template repos (primary + additional)
func (c *Config) GetTemplateRepos() []string {
	repos := []string{c.TemplateRepo}
	for _, r := range c.TemplateRepos {
		if r != c.TemplateRepo {
			repos = append(repos, r)
		}
	}
	return repos
}

// AddTemplateRepo adds a template repo to the list
func (c *Config) AddTemplateRepo(repo string) {
	for _, r := range c.TemplateRepos {
		if r == repo {
			return // already exists
		}
	}
	c.TemplateRepos = append(c.TemplateRepos, repo)
}

// RemoveTemplateRepo removes a template repo from the list
func (c *Config) RemoveTemplateRepo(repo string) {
	filtered := make([]string, 0, len(c.TemplateRepos))
	for _, r := range c.TemplateRepos {
		if r != repo {
			filtered = append(filtered, r)
		}
	}
	c.TemplateRepos = filtered
}

// detectGitHubUser attempts to detect the GitHub user via gh CLI
func detectGitHubUser() string {
	cmd := exec.Command("gh", "api", "user", "-q", ".login")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// String returns a human-readable representation of the config
func (c *Config) String() string {
	data, _ := yaml.Marshal(c)
	return string(data)
}
