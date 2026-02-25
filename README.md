# Sloth Incubator

One command. Fresh project. No finagling.

A CLI/TUI tool for scaffolding new projects with templates, devcontainers, and GitHub integration — built to standardize how you start things.

## Quick Install

```bash
# macOS / Linux
curl -sSL https://raw.githubusercontent.com/HungSloth/sloth-incubator/main/install.sh | bash

# Windows (PowerShell)
iwr -useb https://raw.githubusercontent.com/HungSloth/sloth-incubator/main/install.ps1 | iex
```

Installs the `incubator` binary to `~/.local/bin` (Linux/macOS) or `%LOCALAPPDATA%\Programs\incubator` (Windows).

## Usage

```bash
incubator            # Launch the interactive TUI
incubator new        # Same as above — create a new project
incubator list       # List available templates
incubator version    # Print the installed version
incubator update     # Refresh templates from remote repos
incubator config     # Edit configuration interactively
incubator config --show  # Print current config
incubator add-repo <url> # Add a community template repository
incubator create-template <name> # Create a local template scaffold
incubator preview [project-dir] # Start local noVNC preview
```

### Creating a Project

Run `incubator` (or `incubator new`) and the TUI starts with a main menu:

1. **Create Project** — the existing project scaffolding flow
2. **Create Template** — a local template wizard (base, software, tools)

If you choose **Create Project**, the flow is:

1. **Pick a template** — choose from built-in and remote templates
2. **Answer prompts** — project name, description, visibility, license, whether to create a GitHub repo, and optional preview tooling
3. **Confirm** — review the files that will be created
4. **Scaffold** — project directory is created with all files, a `.devcontainer`, git init, and initial commit
5. **GitHub (optional)** — if you selected repo creation, Incubator creates the remote repo and pushes `HEAD`

Projects are created under `~/projects/` by default (configurable).

## Configuration

Config is stored at `~/.incubator/config.yaml` and is created automatically on first run.

| Field | Default | Description |
|-------|---------|-------------|
| `github_user` | auto-detected via `gh` | GitHub username for repo creation |
| `default_visibility` | `private` | Default repo visibility |
| `default_license` | `MIT` | Default license |
| `project_dir` | `~/projects` | Where new projects are created |
| `local_template_dir` | `~/.incubator/local-templates` | Where local templates are stored |
| `editor` | `none` | Editor to open after scaffolding |
| `template_repo` | `HungSloth/incubator-templates` | Primary remote template repository |
| `template_repos` | `[]` | Additional community template repos |
| `auto_update_check` | `true` | Check for template updates automatically |

## Templates

### Built-in Template

The `empty` template ships with the binary and is always available:

- `README.md` — project readme with name and description
- `CLAUDE.md` — AI development guide
- `.gitignore` — standard ignore file
- `.devcontainer/devcontainer.json` — devcontainer with GitHub CLI, `GH_TOKEN` forwarding, and `gh auth setup-git`

### Remote Templates

Remote templates are fetched from GitHub repos containing a `registry.yaml` at the root. Each template lives in its own directory with a `template.yaml` manifest.

```bash
# Add a community template repo
incubator add-repo owner/repo-name

# Refresh templates
incubator update
```

### Local Template Creator

You can scaffold and maintain templates locally, without publishing a repo first:

```bash
# Create a starter template in ~/.incubator/local-templates/my-template
incubator create-template my-template
```

This generates:

- `template.yaml` (manifest + prompts)
- `files/` (template source files rendered during `incubator new`)
- Ubuntu devcontainer starter with `gh`, `node`, `python`, `go`, and `rust`
- `.incubator/preview/` assets for the built-in virtual display workflow

After creating a local template:

1. Edit `template.yaml` and files under `files/`.
2. Run `incubator list` to confirm it appears.
3. Use it from `incubator` / `incubator new` like any other template.

### Template Creator Wizard (TUI)

From the TUI main menu, choose **Create Template** to launch the interactive wizard.

The wizard creates a local reusable template in `~/.incubator/local-templates/<name>` and guides you through:

1. Template name and description
2. Base preset selection
3. Software preset selection
4. Tool selection (multi-select), including optional devcontainer and preview assets
5. Review and generate

### Template Manifest Format

Templates are defined by a `template.yaml` file:

```yaml
name: my-template
version: "1.0.0"
description: "A project template"
author: "YourName"

prompts:
  - name: project_name
    label: "Project name"
    type: text        # text | select | confirm
    required: true
  - name: language
    label: "Language"
    type: select
    options:
      - label: Go
        value: go
      - label: Python
        value: python
    default: go

files:
  - src: "src/**"
    always: true
  - src: ".docker/**"
    when: "{{if .use_docker}}true{{end}}"

devcontainer:
  base_image: "mcr.microsoft.com/devcontainers/base:ubuntu"
  features:
    always:
      - "ghcr.io/devcontainers/features/github-cli:1"

hooks:
  post_create: "echo 'Setup complete'"
```

Template files use Go's `text/template` syntax. Files ending in `.tmpl` are processed through the template engine (with the `.tmpl` extension stripped from the output). Template variables in directory/file names use `{{variable}}` syntax.

### Headless Preview (Xvfb + noVNC)

If you enable preview tooling during project creation, Incubator generates preview assets in `.incubator/preview/`.

Run:

```bash
incubator preview .
```

This command:
- loads `.incubator/preview/config.yaml`
- builds a local Docker image for preview
- starts Xvfb + x11vnc + noVNC in a container
- opens `http://localhost:<novnc_port>` in your default browser

Notes:
- Docker is required on the host machine.
- `app_command` in `.incubator/preview/config.yaml` controls what app/process runs in the virtual display.

## Dev Container Auth

Generated projects include a devcontainer configured for seamless GitHub auth. For this to work, set up your host machine once:

```bash
# Add to ~/.zshrc (or ~/.bashrc)
export GH_TOKEN=$(gh auth token)
```

Then launch Cursor/VS Code from a terminal so it inherits the env var. The devcontainer's `remoteEnv` forwards `GH_TOKEN` into the container, and `gh auth setup-git` runs on creation so both `gh` and `git push` work without additional login.

## Development

### Prerequisites

- Go 1.22+
- `gh` CLI (for GitHub operations)

### Build from Source

```bash
git clone https://github.com/HungSloth/sloth-incubator.git
cd sloth-incubator
go build -o incubator ./cmd/incubator
./incubator version
```

### Dev Container

The repo includes a `.devcontainer/` for development. It provides Ubuntu with GitHub CLI, Node, and Python pre-installed.

### Project Structure

```
cmd/incubator/main.go          CLI entrypoint (Cobra commands)
internal/
  tui/                          Bubbletea TUI models
    app.go                        Main app model and navigation
    picker.go                     Template picker
    form.go                       Prompt form
    confirm.go                    Confirmation screen
    progress.go                   Scaffolding progress
    done.go                       Completion screen
  template/                     Template engine
    manifest.go                   Manifest types (prompts, file rules, devcontainer)
    embedded.go                   Built-in template embed
    embedded/empty/               Built-in "empty" template files
    renderer.go                   Template rendering (file walking, Go templates)
    loader.go                     Remote template fetching and registry
    cache.go                      Template cache management
    hooks.go                      Post-create hooks
    registry_remote.go            Remote registry support
  git/                          Git and GitHub operations
  config/                       User configuration (~/.incubator/config.yaml)
  updater/                      Self-update (future)
.goreleaser.yml                 Cross-platform release config
install.sh                      One-line installer (macOS/Linux)
install.ps1                     One-line installer (Windows)
```

### Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/lipgloss` | TUI styling |
| `github.com/charmbracelet/bubbles` | TUI components (text input, lists) |
| `github.com/spf13/cobra` | CLI framework |
| `gopkg.in/yaml.v3` | YAML parsing |

### Testing

```bash
go test ./...
```

### Creating a Release

Releases are built with [GoReleaser](https://goreleaser.com) and produce binaries for linux/darwin/windows on amd64/arm64.

```bash
# Tag the version (this triggers automated GitHub release workflow)
git tag v0.x.0
git push origin v0.x.0
```

The `Release` GitHub Actions workflow runs automatically on version tags (`v*`) and publishes release artifacts.

For local dry-runs:

```bash
goreleaser release --snapshot --clean
```

## License

MIT
