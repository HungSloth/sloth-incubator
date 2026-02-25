package template

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var localTemplateNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// LocalTemplateWizardOptions captures wizard selections for local template generation.
type LocalTemplateWizardOptions struct {
	Name        string
	Description string
	Base        string
	Software    string
	Tools       []string
}

// CreateLocalTemplate scaffolds a local template directory with starter files.
func CreateLocalTemplate(localTemplatesDir, name string) (string, error) {
	if !localTemplateNamePattern.MatchString(name) {
		return "", fmt.Errorf("invalid template name %q: use letters, numbers, dashes, and underscores", name)
	}

	templateDir := filepath.Join(localTemplatesDir, name)
	if _, err := os.Stat(templateDir); err == nil {
		return "", fmt.Errorf("template %q already exists at %s", name, templateDir)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("checking template path: %w", err)
	}

	for relPath, content := range localTemplateScaffold(name) {
		fullPath := filepath.Join(templateDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return "", fmt.Errorf("creating directory for %s: %w", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("writing %s: %w", relPath, err)
		}
	}

	return templateDir, nil
}

// CreateLocalTemplateFromWizard scaffolds a local template from wizard options.
func CreateLocalTemplateFromWizard(localTemplatesDir string, opts LocalTemplateWizardOptions) (string, error) {
	name := strings.TrimSpace(opts.Name)
	if !localTemplateNamePattern.MatchString(name) {
		return "", fmt.Errorf("invalid template name %q: use letters, numbers, dashes, and underscores", name)
	}

	if strings.TrimSpace(opts.Description) == "" {
		opts.Description = fmt.Sprintf("Local template scaffold for %s", name)
	}
	opts.Name = name
	if opts.Base == "" {
		opts.Base = "empty"
	}
	if opts.Software == "" {
		opts.Software = "go"
	}

	templateDir := filepath.Join(localTemplatesDir, name)
	if _, err := os.Stat(templateDir); err == nil {
		return "", fmt.Errorf("template %q already exists at %s", name, templateDir)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("checking template path: %w", err)
	}

	for relPath, content := range localTemplateScaffoldWithOptions(opts) {
		fullPath := filepath.Join(templateDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return "", fmt.Errorf("creating directory for %s: %w", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("writing %s: %w", relPath, err)
		}
	}

	return templateDir, nil
}

// LoadLocalManifests loads template manifests from a local template directory.
func LoadLocalManifests(localTemplatesDir string) ([]*TemplateManifest, error) {
	entries, err := os.ReadDir(localTemplatesDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading local templates directory: %w", err)
	}

	manifests := make([]*TemplateManifest, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		templateDir := filepath.Join(localTemplatesDir, entry.Name())
		manifest, err := loadLocalManifest(templateDir)
		if err != nil {
			// Skip invalid local templates and continue loading others.
			continue
		}
		manifests = append(manifests, manifest)
	}

	return manifests, nil
}

func loadLocalManifest(templateDir string) (*TemplateManifest, error) {
	manifestPath := filepath.Join(templateDir, "template.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	filesDir := filepath.Join(templateDir, "files")
	stat, err := os.Stat(filesDir)
	if err != nil || !stat.IsDir() {
		return nil, fmt.Errorf("missing files directory at %s", filesDir)
	}

	var manifest TemplateManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	if manifest.Name == "" {
		manifest.Name = filepath.Base(templateDir)
	}
	absDir, err := filepath.Abs(templateDir)
	if err != nil {
		return nil, err
	}
	manifest.SourcePath = absDir
	manifest.IsBuiltin = false
	manifest.ApplyDefaults()

	return &manifest, nil
}

func localTemplateScaffold(name string) map[string]string {
	return map[string]string{
		"template.yaml": localTemplateYAML(name),
		"files/README.md.tmpl": `# {{.project_name}}

{{.description}}
`,
		"files/.gitignore": `.DS_Store
.env
`,
		"files/.devcontainer/devcontainer.json.tmpl": `{
  "name": "{{.project_name}}",
  "image": "mcr.microsoft.com/devcontainers/base:ubuntu",
  "features": {
    "ghcr.io/devcontainers/features/github-cli:1": {},
    "ghcr.io/devcontainers/features/node:1": {},
    "ghcr.io/devcontainers/features/python:1": {},
    "ghcr.io/devcontainers/features/go:1": {},
    "ghcr.io/devcontainers/features/rust:1": {}
  },
  "postCreateCommand": "gh auth setup-git && echo 'Local template dev container ready!'",
  "remoteEnv": {
    "GH_TOKEN": "${localEnv:GH_TOKEN}"
  },
  "customizations": {
    "vscode": {
      "settings": {},
      "extensions": []
    }
  }
}
`,
		"files/.incubator/preview/config.yaml.tmpl": previewConfigTemplate,
		"files/.incubator/preview/entrypoint.sh":    previewEntrypointTemplate,
		"files/.incubator/preview/Dockerfile":       previewDockerfileTemplate,
	}
}

func localTemplateScaffoldWithOptions(opts LocalTemplateWizardOptions) map[string]string {
	files := map[string]string{
		"template.yaml":        localTemplateYAMLWithOptions(opts),
		"files/README.md.tmpl": readmeTemplateForOptions(opts),
		"files/.gitignore": `.DS_Store
.env
`,
	}

	if !hasTool(opts.Tools, "git") {
		delete(files, "files/.gitignore")
	}

	if hasTool(opts.Tools, "devcontainer") {
		files["files/.devcontainer/devcontainer.json.tmpl"] = devcontainerTemplateForSoftware(opts.Software)
	}

	if hasTool(opts.Tools, "preview") {
		files["files/.incubator/preview/config.yaml.tmpl"] = previewConfigTemplate
		files["files/.incubator/preview/entrypoint.sh"] = previewEntrypointTemplate
		files["files/.incubator/preview/Dockerfile"] = previewDockerfileTemplate
	}

	return files
}

func localTemplateYAML(name string) string {
	return fmt.Sprintf(`name: %s
version: 0.1.0
description: Local template scaffold for %s
author: local
prompts:
  - name: project_name
    label: Project name
    type: text
    required: true
  - name: description
    label: Description
    type: text
    default: A project created from a local template
  - name: visibility
    label: Repo visibility
    type: select
    options:
      - private
      - public
    default: private
  - name: license
    label: License
    type: select
    options:
      - MIT
      - Apache-2.0
      - GPL-3.0
      - none
    default: MIT
  - name: create_github_repo
    label: Create GitHub repository?
    type: confirm
    default: true
  - name: enable_preview
    label: Enable headless preview tooling?
    type: confirm
    default: true
files:
  - src: .incubator/preview/**
    when: '{{if .enable_preview}}true{{end}}'
preview:
  enabled: true
  app_command: "echo \"Set preview app_command in .incubator/preview/config.yaml\" && sleep infinity"
devcontainer:
  base_image: ubuntu
  features:
    always:
      - gh
      - node
      - python
      - go
      - rust
      - incubator-preview
`, name, name)
}

func localTemplateYAMLWithOptions(opts LocalTemplateWizardOptions) string {
	var b strings.Builder

	fmt.Fprintf(&b, "name: %s\n", opts.Name)
	b.WriteString("version: 0.1.0\n")
	fmt.Fprintf(&b, "description: %q\n", opts.Description)
	b.WriteString("author: local\n")
	b.WriteString("prompts:\n")
	b.WriteString("  - name: project_name\n")
	b.WriteString("    label: Project name\n")
	b.WriteString("    type: text\n")
	b.WriteString("    required: true\n")
	b.WriteString("  - name: description\n")
	b.WriteString("    label: Description\n")
	b.WriteString("    type: text\n")
	b.WriteString("    default: A project created from a local template\n")
	b.WriteString("  - name: create_github_repo\n")
	b.WriteString("    label: Create GitHub repository?\n")
	b.WriteString("    type: confirm\n")
	b.WriteString("    default: true\n")

	if hasTool(opts.Tools, "preview") {
		b.WriteString("  - name: enable_preview\n")
		b.WriteString("    label: Enable headless preview tooling?\n")
		b.WriteString("    type: confirm\n")
		b.WriteString("    default: true\n")
		b.WriteString("files:\n")
		b.WriteString("  - src: .incubator/preview/**\n")
		b.WriteString("    when: '{{if .enable_preview}}true{{end}}'\n")
		b.WriteString("preview:\n")
		b.WriteString("  enabled: true\n")
		b.WriteString("  app_command: \"echo \\\"Set preview app_command in .incubator/preview/config.yaml\\\" && sleep infinity\"\n")
	}

	if hasTool(opts.Tools, "devcontainer") {
		b.WriteString("devcontainer:\n")
		b.WriteString("  base_image: ubuntu\n")
		b.WriteString("  features:\n")
		b.WriteString("    always:\n")
		for _, feature := range devcontainerFeaturesForSoftware(opts.Software) {
			fmt.Fprintf(&b, "      - %s\n", feature)
		}
	}

	// Preserve wizard metadata in comments so users can trace generated defaults.
	fmt.Fprintf(&b, "# wizard_base: %s\n", opts.Base)
	fmt.Fprintf(&b, "# wizard_software: %s\n", opts.Software)

	return b.String()
}

func readmeTemplateForOptions(opts LocalTemplateWizardOptions) string {
	var b strings.Builder
	b.WriteString("# {{.project_name}}\n\n")
	b.WriteString("{{.description}}\n\n")
	b.WriteString(fmt.Sprintf("Generated with base `%s` and software preset `%s`.\n", opts.Base, opts.Software))
	return b.String()
}

func devcontainerTemplateForSoftware(software string) string {
	features := devcontainerFeaturesForSoftware(software)
	var featureEntries strings.Builder
	for _, feature := range features {
		fmt.Fprintf(&featureEntries, "    \"%s\": {},\n", featureToDevcontainerRef(feature))
	}

	return fmt.Sprintf(`{
  "name": "{{.project_name}}",
  "image": "mcr.microsoft.com/devcontainers/base:ubuntu",
  "features": {
%s  },
  "postCreateCommand": "gh auth setup-git && echo 'Local template dev container ready!'",
  "remoteEnv": {
    "GH_TOKEN": "${localEnv:GH_TOKEN}"
  },
  "customizations": {
    "vscode": {
      "settings": {},
      "extensions": []
    }
  }
}
`, featureEntries.String())
}

func devcontainerFeaturesForSoftware(software string) []string {
	base := []string{"gh"}
	switch software {
	case "node":
		return append(base, "node")
	case "python":
		return append(base, "python")
	default:
		return append(base, "go")
	}
}

func featureToDevcontainerRef(feature string) string {
	switch feature {
	case "gh":
		return "ghcr.io/devcontainers/features/github-cli:1"
	case "node":
		return "ghcr.io/devcontainers/features/node:1"
	case "python":
		return "ghcr.io/devcontainers/features/python:1"
	case "go":
		return "ghcr.io/devcontainers/features/go:1"
	default:
		return feature
	}
}

func hasTool(tools []string, tool string) bool {
	for _, t := range tools {
		if t == tool {
			return true
		}
	}
	return false
}

const previewConfigTemplate = `enabled: true
app_command: "echo \"Set preview app_command in .incubator/preview/config.yaml\" && sleep infinity"
novnc_port: 6080
vnc_port: 5900
`

const previewEntrypointTemplate = `#!/usr/bin/env bash
set -euo pipefail

DISPLAY_NUM="${DISPLAY_NUM:-:99}"
SCREEN_GEOMETRY="${SCREEN_GEOMETRY:-1280x800x24}"
VNC_PORT="${VNC_PORT:-5900}"
NOVNC_PORT="${NOVNC_PORT:-6080}"
APP_COMMAND="${PREVIEW_APP_COMMAND:-}"

Xvfb "${DISPLAY_NUM}" -screen 0 "${SCREEN_GEOMETRY}" &
XVFB_PID=$!

export DISPLAY="${DISPLAY_NUM}"
openbox >/tmp/openbox.log 2>&1 &

x11vnc -display "${DISPLAY}" -rfbport "${VNC_PORT}" -forever -shared -nopw -listen 0.0.0.0 >/tmp/x11vnc.log 2>&1 &
xterm >/tmp/xterm.log 2>&1 &

if [[ -n "${APP_COMMAND}" ]]; then
  bash -lc "${APP_COMMAND}" >/tmp/app.log 2>&1 &
fi

NOVNC_DIR="/usr/share/novnc"
if [[ ! -d "${NOVNC_DIR}" ]]; then
  NOVNC_DIR="/usr/share/novnc/utils/novnc_proxy"
fi

if [[ -x /usr/share/novnc/utils/novnc_proxy ]]; then
  /usr/share/novnc/utils/novnc_proxy --listen "${NOVNC_PORT}" --vnc "127.0.0.1:${VNC_PORT}" >/tmp/novnc.log 2>&1 &
else
  websockify --web "${NOVNC_DIR}" "${NOVNC_PORT}" "127.0.0.1:${VNC_PORT}" >/tmp/novnc.log 2>&1 &
fi

wait "${XVFB_PID}"
`

const previewDockerfileTemplate = `FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    xvfb \
    x11vnc \
    novnc \
    websockify \
    openbox \
    xterm \
    ca-certificates \
    bash \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
COPY .incubator/preview/entrypoint.sh /usr/local/bin/preview-entrypoint.sh
RUN chmod +x /usr/local/bin/preview-entrypoint.sh

EXPOSE 5900 6080

ENTRYPOINT ["/usr/local/bin/preview-entrypoint.sh"]
`
