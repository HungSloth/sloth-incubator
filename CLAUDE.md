# Sloth Incubator — Development Guide

## Project Overview

Sloth Incubator is a CLI/TUI tool for scaffolding new projects. Built with Go, Bubbletea (TUI), and Cobra (CLI).

## Build & Run

```bash
# Build
go build -o incubator ./cmd/incubator

# Run
./incubator          # Launch TUI
./incubator new      # Create a new project
./incubator list     # List templates
./incubator version  # Print version
./incubator update   # Refresh templates
./incubator config   # Edit config interactively (--show to print)
./incubator add-repo <url>  # Add a community template repo
```

## Install (one-liner)

```bash
# macOS / Linux
curl -sSL https://raw.githubusercontent.com/HungSloth/sloth-incubator/main/install.sh | bash

# Windows (PowerShell)
iwr -useb https://raw.githubusercontent.com/HungSloth/sloth-incubator/main/install.ps1 | iex
```

## Project Structure

- `cmd/incubator/main.go` — CLI entrypoint with Cobra commands
- `internal/tui/` — Bubbletea TUI models (app, picker, form, confirm, progress, done)
- `internal/template/` — Template engine, manifest types, embedded templates
- `internal/template/embedded/` — Built-in template files (use `all:` embed prefix for dotfiles)
- `internal/git/` — Git and GitHub operations
- `internal/config/` — User configuration
- `internal/updater/` — Self-update (future)
- `.goreleaser.yml` — Cross-platform release config (linux/darwin/windows, amd64/arm64)
- `.devcontainer/` — Dev container config with GH_TOKEN forwarding

## Key Dependencies

- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — TUI styling
- `github.com/charmbracelet/bubbles` — TUI components
- `github.com/spf13/cobra` — CLI framework
- `gopkg.in/yaml.v3` — YAML parsing

## Testing

```bash
go test ./...
```

## Release

Uses GoReleaser. Tag a version and run goreleaser:
```bash
git tag v0.x.0
git push origin v0.x.0
GITHUB_TOKEN=$GH_TOKEN goreleaser release --clean
```

## Dev Container Auth

The devcontainer forwards `GH_TOKEN` from the host via `remoteEnv` and runs
`gh auth setup-git` on creation. To set up on the host machine:
```bash
echo 'export GH_TOKEN=$(gh auth token)' >> ~/.zshrc
```
Launch Cursor from a terminal so it inherits the env var.

## Embedded Templates

Template files under `internal/template/embedded/` are compiled into the binary
via `//go:embed all:embedded/empty`. The `all:` prefix is required to include
dotfiles like `.devcontainer/` and `.gitignore`.
