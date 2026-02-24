package template

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// RegistryEntry represents an entry in registry.yaml
type RegistryEntry struct {
	Name        string `yaml:"name"`
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
}

// Registry represents the registry.yaml file
type Registry struct {
	Templates []RegistryEntry `yaml:"templates"`
}

// Loader handles loading templates from various sources
type Loader struct {
	cacheDir     string
	templateRepo string
}

// NewLoader creates a new template loader
func NewLoader(cacheDir, templateRepo string) *Loader {
	return &Loader{
		cacheDir:     cacheDir,
		templateRepo: templateRepo,
	}
}

// TemplatesDir returns the path to the cached templates directory
func (l *Loader) TemplatesDir() string {
	return filepath.Join(l.cacheDir, "templates")
}

// FetchTemplates clones or pulls the templates repository
func (l *Loader) FetchTemplates() error {
	templatesDir := l.TemplatesDir()

	if _, err := os.Stat(filepath.Join(templatesDir, ".git")); err == nil {
		// Already cloned, pull
		return l.pullTemplates(templatesDir)
	}

	// Fresh clone
	return l.cloneTemplates(templatesDir)
}

// cloneTemplates performs a shallow clone of the templates repo
func (l *Loader) cloneTemplates(dir string) error {
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	repoURL := fmt.Sprintf("https://github.com/%s.git", l.templateRepo)
	cmd := exec.Command("git", "clone", "--depth=1", repoURL, dir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cloning templates repo: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// pullTemplates pulls the latest changes
func (l *Loader) pullTemplates(dir string) error {
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pulling templates: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// LoadRegistry loads and parses the registry.yaml file
func (l *Loader) LoadRegistry() (*Registry, error) {
	registryPath := filepath.Join(l.TemplatesDir(), "registry.yaml")
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil, fmt.Errorf("reading registry.yaml: %w", err)
	}

	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parsing registry.yaml: %w", err)
	}

	return &reg, nil
}

// LoadManifest loads a template manifest from the templates directory
func (l *Loader) LoadManifest(templatePath string) (*TemplateManifest, error) {
	manifestPath := filepath.Join(l.TemplatesDir(), templatePath, "template.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("reading template.yaml: %w", err)
	}

	var manifest TemplateManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parsing template.yaml: %w", err)
	}

	return &manifest, nil
}

// LoadAllManifests loads all template manifests from the registry
func (l *Loader) LoadAllManifests() ([]*TemplateManifest, error) {
	reg, err := l.LoadRegistry()
	if err != nil {
		return nil, err
	}

	var manifests []*TemplateManifest
	for _, entry := range reg.Templates {
		manifest, err := l.LoadManifest(entry.Path)
		if err != nil {
			// Skip templates that fail to load
			continue
		}
		manifests = append(manifests, manifest)
	}

	return manifests, nil
}

// GetTemplateFS returns an os.DirFS for the template directory
func (l *Loader) GetTemplateFS(templatePath string) (string, error) {
	dir := filepath.Join(l.TemplatesDir(), templatePath)
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("template directory not found: %s", dir)
	}
	return dir, nil
}
