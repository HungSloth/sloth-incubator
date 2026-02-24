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
```

## Project Structure

- `cmd/incubator/main.go` — CLI entrypoint with Cobra commands
- `internal/tui/` — Bubbletea TUI models (app, picker, form, confirm, progress, done)
- `internal/template/` — Template engine, manifest types, embedded templates
- `internal/git/` — Git and GitHub operations
- `internal/config/` — User configuration (future)
- `internal/updater/` — Self-update (future)

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

Uses GoReleaser. Tag a version to trigger a release:
```bash
git tag v0.1.0
git push --tags
```
