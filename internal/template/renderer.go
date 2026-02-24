package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Renderer handles template rendering
type Renderer struct {
	manifest *TemplateManifest
	answers  map[string]interface{}
}

// NewRenderer creates a new template renderer
func NewRenderer(manifest *TemplateManifest, answers map[string]interface{}) *Renderer {
	return &Renderer{
		manifest: manifest,
		answers:  answers,
	}
}

// RenderTo renders the template to the target directory using the provided filesystem
func (r *Renderer) RenderTo(targetDir string, sourceFS fs.FS) error {
	return fs.WalkDir(sourceFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root
		if path == "." {
			return nil
		}

		// Expand template variables in path names
		expandedPath := r.expandPath(path)
		targetPath := filepath.Join(targetDir, expandedPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Read source file
		content, err := fs.ReadFile(sourceFS, path)
		if err != nil {
			return fmt.Errorf("reading template file %s: %w", path, err)
		}

		// Process .tmpl files through text/template
		if strings.HasSuffix(path, ".tmpl") {
			targetPath = strings.TrimSuffix(targetPath, ".tmpl")
			processed, err := r.processTemplate(path, string(content))
			if err != nil {
				return fmt.Errorf("processing template %s: %w", path, err)
			}
			content = []byte(processed)
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		return os.WriteFile(targetPath, content, 0644)
	})
}

// expandPath replaces {{variable}} patterns in path names
func (r *Renderer) expandPath(path string) string {
	result := path
	for key, val := range r.answers {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", val))
	}
	// Also remove .tmpl suffix from directory traversal
	return result
}

// processTemplate processes a single template file
func (r *Renderer) processTemplate(name, content string) (string, error) {
	tmpl, err := template.New(name).Option("missingkey=zero").Parse(content)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, r.answers); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ListFiles returns the list of files that would be created
func (r *Renderer) ListFiles(sourceFS fs.FS) ([]string, error) {
	var files []string
	err := fs.WalkDir(sourceFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." || d.IsDir() {
			return nil
		}
		expandedPath := r.expandPath(path)
		if strings.HasSuffix(expandedPath, ".tmpl") {
			expandedPath = strings.TrimSuffix(expandedPath, ".tmpl")
		}
		files = append(files, expandedPath)
		return nil
	})
	return files, err
}
