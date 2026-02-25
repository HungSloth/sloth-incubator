package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateLocalTemplateAndLoadManifests(t *testing.T) {
	root := t.TempDir()

	templateDir, err := CreateLocalTemplate(root, "demo_template")
	if err != nil {
		t.Fatalf("CreateLocalTemplate returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(templateDir, "template.yaml")); err != nil {
		t.Fatalf("expected template.yaml to exist: %v", err)
	}
	if stat, err := os.Stat(filepath.Join(templateDir, "files")); err != nil || !stat.IsDir() {
		t.Fatalf("expected files directory to exist")
	}

	manifests, err := LoadLocalManifests(root)
	if err != nil {
		t.Fatalf("LoadLocalManifests returned error: %v", err)
	}
	if len(manifests) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(manifests))
	}
	if manifests[0].Name != "demo_template" {
		t.Fatalf("expected manifest name demo_template, got %s", manifests[0].Name)
	}
	if manifests[0].SourcePath != templateDir {
		t.Fatalf("expected SourcePath %s, got %s", templateDir, manifests[0].SourcePath)
	}
}

func TestResolveTemplateFSLocalTemplate(t *testing.T) {
	root := t.TempDir()
	if _, err := CreateLocalTemplate(root, "demo"); err != nil {
		t.Fatalf("CreateLocalTemplate returned error: %v", err)
	}

	manifests, err := LoadLocalManifests(root)
	if err != nil {
		t.Fatalf("LoadLocalManifests returned error: %v", err)
	}
	if len(manifests) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(manifests))
	}

	templateFS, err := ResolveTemplateFS(manifests[0], t.TempDir(), "owner/repo")
	if err != nil {
		t.Fatalf("ResolveTemplateFS returned error: %v", err)
	}
	content, err := fs.ReadFile(templateFS, "README.md.tmpl")
	if err != nil {
		t.Fatalf("expected README.md.tmpl in template FS: %v", err)
	}
	if len(content) == 0 {
		t.Fatalf("expected README.md.tmpl to have content")
	}
}

func TestCreateLocalTemplateRejectsInvalidName(t *testing.T) {
	if _, err := CreateLocalTemplate(t.TempDir(), "bad/name"); err == nil {
		t.Fatalf("expected invalid template name error")
	}
}

func TestRenderLocalTemplate(t *testing.T) {
	root := t.TempDir()
	if _, err := CreateLocalTemplate(root, "demo"); err != nil {
		t.Fatalf("CreateLocalTemplate returned error: %v", err)
	}

	manifests, err := LoadLocalManifests(root)
	if err != nil {
		t.Fatalf("LoadLocalManifests returned error: %v", err)
	}
	answers := map[string]interface{}{
		"project_name":       "hello-local",
		"description":        "local template output",
		"create_github_repo": false,
		"enable_preview":     true,
	}
	templateFS, err := ResolveTemplateFS(manifests[0], t.TempDir(), "owner/repo")
	if err != nil {
		t.Fatalf("ResolveTemplateFS returned error: %v", err)
	}
	outputDir := t.TempDir()
	renderer := NewRenderer(manifests[0], answers)
	if err := renderer.RenderTo(outputDir, templateFS); err != nil {
		t.Fatalf("RenderTo returned error: %v", err)
	}

	readmePath := filepath.Join(outputDir, "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("expected rendered README.md to exist: %v", err)
	}
	if string(content) == "" {
		t.Fatalf("expected rendered README.md to have content")
	}
}
