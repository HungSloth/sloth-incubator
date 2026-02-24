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

		// Check conditional file inclusion
		if !r.shouldInclude(path) {
			if d.IsDir() {
				return fs.SkipDir
			}
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

// shouldInclude checks if a file/directory should be included based on manifest rules
func (r *Renderer) shouldInclude(path string) bool {
	// If no file rules defined, include everything
	if len(r.manifest.Files) == 0 {
		return true
	}

	for _, rule := range r.manifest.Files {
		if rule.Always {
			if matchesGlob(rule.Src, path) {
				return true
			}
			continue
		}

		if rule.When != "" {
			// Evaluate the when condition as a Go template
			if r.evaluateCondition(rule.When) && matchesGlob(rule.Src, path) {
				return true
			}
		}
	}

	// If rules are defined but none match, still include files not covered by any rule
	// This handles the case where rules only cover specific directories
	for _, rule := range r.manifest.Files {
		expandedSrc := r.expandPath(rule.Src)
		if matchesGlob(expandedSrc, path) {
			return false // A rule covers this path but conditions weren't met
		}
	}

	return true // No rule covers this path, include by default
}

// evaluateCondition evaluates a when condition template expression
func (r *Renderer) evaluateCondition(condition string) bool {
	tmpl, err := template.New("condition").Option("missingkey=zero").Parse(condition)
	if err != nil {
		return false
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, r.answers); err != nil {
		return false
	}

	result := strings.TrimSpace(buf.String())
	return result != "" && result != "false" && result != "0" && result != "<no value>"
}

// matchesGlob checks if a path matches a glob pattern
func matchesGlob(pattern, path string) bool {
	// Handle ** glob patterns
	if strings.Contains(pattern, "**") {
		prefix := strings.Split(pattern, "**")[0]
		return strings.HasPrefix(path, prefix)
	}

	matched, _ := filepath.Match(pattern, path)
	if matched {
		return true
	}

	// Also check if the path is within the pattern's directory
	dir := filepath.Dir(pattern)
	return strings.HasPrefix(path, dir+"/")
}

// expandPath replaces {{variable}} patterns in path names
func (r *Renderer) expandPath(path string) string {
	result := path
	for key, val := range r.answers {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", val))
	}
	return result
}

// processTemplate processes a single template file
func (r *Renderer) processTemplate(name, content string) (string, error) {
	funcMap := template.FuncMap{
		"ne": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b)
		},
		"eq": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
		},
	}

	tmpl, err := template.New(name).Funcs(funcMap).Option("missingkey=zero").Parse(content)
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
		if !r.shouldInclude(path) {
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
