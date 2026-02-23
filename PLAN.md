# Sloth Incubator — Project Plan

> One command. Fresh project. No finagling.

## Vision

A CLI/TUI tool that standardizes how projects are created. Install with a one-liner, pick a template, answer a few questions, and walk away with a fully scaffolded project — git repo, GitHub remote, devcontainer, and all the boilerplate handled.

Templates live in a separate repo so they can evolve independently from the tool itself.

---

## Architecture

### Components

```
┌─────────────────────────────────────────────┐
│  sloth-incubator (CLI/TUI binary)           │
│  ┌───────────┐  ┌──────────┐  ┌──────────┐ │
│  │ TUI Layer  │  │ Template │  │ Git/GH   │ │
│  │ Bubbletea  │  │ Engine   │  │ Provider │ │
│  └───────────┘  └──────────┘  └──────────┘ │
└──────────────────────┬──────────────────────┘
                       │ fetches/caches
          ┌────────────▼────────────┐
          │  incubator-templates    │
          │  (separate repo)        │
          │  ├── node-api/          │
          │  ├── python-cli/        │
          │  ├── go-tool/           │
          │  └── ...                │
          └─────────────────────────┘
```

### Tech Stack

| Component | Choice | Why |
|-----------|--------|-----|
| Language | Go | Single binary, cross-platform, no runtime |
| TUI | Bubbletea + Lipgloss + Bubbles | Best-in-class terminal UI (Charm stack) |
| Templates | Go `text/template` | Native, no extra deps |
| Git ops | go-git + os/exec for `gh` | In-process git, shell out for GitHub |
| Config | YAML | Human-friendly template manifests |
| Updates | GitHub Releases API | Self-update from releases |

### Platform Support

- macOS (arm64 + amd64)
- Linux (arm64 + amd64)
- Windows (amd64) — stretch goal

---

## Installation

### One-liner

```bash
curl -sSL https://raw.githubusercontent.com/HungSloth/sloth-incubator/main/install.sh | bash
```

The installer:
1. Detects OS and architecture
2. Downloads the latest release binary from GitHub Releases
3. Places it in `~/.local/bin/incubator` (or `/usr/local/bin` with sudo)
4. Adds to PATH if needed
5. Verifies with `incubator --version`

### Manual

Download from GitHub Releases, put binary in PATH.

---

## CLI Interface

```
incubator              # Launch TUI (default)
incubator new          # Launch TUI in create mode
incubator list         # List available templates
incubator update       # Update incubator + refresh templates
incubator config       # Edit settings (GitHub user, default visibility, etc.)
incubator version      # Print version
```

---

## TUI Flow

All TUI screens use the **alternate buffer** (clean entry/exit).

### Screen 1: Template Selection

```
╔══════════════════════════════════════════╗
║  🦥  Sloth Incubator                    ║
╚══════════════════════════════════════════╝

  Search: _

  > ● node-api         Node.js API (Express/Fastify/Hono)
    ○ python-cli       Python CLI with Click/Typer
    ○ go-tool          Go CLI tool with Cobra
    ○ react-app        React + Vite frontend
    ○ static-site      HTML/CSS/JS static site
    ○ empty            Blank project with devcontainer

  ↑/↓ navigate  /  search  enter select  q quit
```

### Screen 2: Configuration

Dynamic prompts based on `template.yaml`:

```
╔══════════════════════════════════════════╗
║  node-api — Configure                   ║
╚══════════════════════════════════════════╝

  Project name:  _______________
  Description:   _______________
  Framework:     [ Express ▾ ]
  Database:      [ None ▾ ]
  Visibility:    ◉ Private  ○ Public
  License:       [ MIT ▾ ]
  Docker:        [✓]
  CI (Actions):  [✓]

  tab next  shift+tab prev  enter confirm
```

### Screen 3: Confirmation

```
╔══════════════════════════════════════════╗
║  Ready to create                        ║
╚══════════════════════════════════════════╝

  Name:        my-cool-api
  Template:    node-api (Express + Postgres)
  Visibility:  private
  License:     MIT
  Extras:      Docker, GitHub Actions CI

  Files to create:
    .devcontainer/devcontainer.json
    src/index.ts
    package.json
    Dockerfile
    .github/workflows/ci.yml
    CLAUDE.md
    README.md
    .gitignore
    .env.example

  enter create  esc back  q cancel
```

### Screen 4: Progress

```
  ✓ Creating project directory
  ✓ Rendering templates
  ✓ Initializing git repo
  ● Creating GitHub repo...
  ○ Pushing to origin

  [████████████░░░░░░░░] 60%
```

### Screen 5: Done

```
  ✨ my-cool-api is ready!

  Local:   ~/projects/my-cool-api
  Remote:  https://github.com/HungSloth/my-cool-api

  cd ~/projects/my-cool-api

  enter open in editor  q exit
```

---

## Template Format

Templates live in the `incubator-templates` repo. Each template is a directory:

```
incubator-templates/
├── registry.yaml           # master index
├── node-api/
│   ├── template.yaml       # manifest
│   ├── {{name}}/           # templated directory names
│   │   ├── package.json.tmpl
│   │   ├── src/
│   │   │   └── index.ts.tmpl
│   │   ├── .devcontainer/
│   │   │   └── devcontainer.json.tmpl
│   │   ├── CLAUDE.md.tmpl
│   │   └── .gitignore
│   └── hooks/
│       └── post_create.sh
├── python-cli/
│   ├── template.yaml
│   └── ...
└── ...
```

### template.yaml

```yaml
name: node-api
version: 1.0.0
description: Node.js API with framework choice
author: HungSloth

prompts:
  - name: project_name
    label: Project name
    type: text
    validate: kebab-case
    required: true

  - name: description
    label: Description
    type: text
    default: "A new API project"

  - name: framework
    label: Framework
    type: select
    options:
      - label: Express
        value: express
      - label: Fastify
        value: fastify
      - label: Hono
        value: hono

  - name: database
    label: Database
    type: select
    options:
      - label: PostgreSQL
        value: postgres
      - label: SQLite
        value: sqlite
      - label: None
        value: none

  - name: docker
    label: Include Dockerfile
    type: confirm
    default: true

  - name: ci
    label: GitHub Actions CI
    type: confirm
    default: true

  - name: visibility
    label: Repo visibility
    type: select
    options: [public, private]
    default: private

  - name: license
    label: License
    type: select
    options: [MIT, Apache-2.0, GPL-3.0, none]
    default: MIT

# Conditional file inclusion
files:
  - src: base/**
    always: true
  - src: docker/**
    when: "{{ .docker }}"
  - src: ci/**
    when: "{{ .ci }}"
  - src: db/{{ .database }}/**
    when: "{{ ne .database \"none\" }}"

# Devcontainer features to include based on selections
devcontainer:
  base_image: mcr.microsoft.com/devcontainers/base:ubuntu
  features:
    always:
      - ghcr.io/devcontainers/features/github-cli:1
      - ghcr.io/devcontainers/features/node:1
    when_database_postgres:
      - ghcr.io/devcontainers-contrib/features/postgres-asdf:1

hooks:
  post_create: hooks/post_create.sh
```

---

## Template Caching

- Templates cached in `~/.incubator/templates/`
- On launch: check last fetch time → if >24h, pull latest in background
- `incubator update` forces refresh
- Cache stores git shallow clone for efficiency
- Offline mode works from cache

---

## Auto-Update

- On launch: background goroutine checks GitHub Releases API
- If newer version: subtle banner at bottom of TUI
  ```
  ⬆ Update available: v1.2.0 → v1.3.0  (run incubator update)
  ```
- `incubator update`: downloads new binary, replaces self, refreshes templates
- No forced updates, never blocks the user

---

## Configuration

Stored in `~/.incubator/config.yaml`:

```yaml
github_user: HungSloth
default_visibility: private
default_license: MIT
project_dir: ~/projects          # where new projects go
editor: cursor                   # open after create (cursor, code, vim, none)
template_repo: HungSloth/incubator-templates
auto_update_check: true
```

---

## Milestones

### v0.1 — Walking (MVP)
- [ ] Go project scaffolding (go mod, main.go)
- [ ] Basic TUI: template list → prompts → create
- [ ] Single built-in template (empty project with devcontainer)
- [ ] Git init + GitHub repo creation via `gh`
- [ ] Install script (curl one-liner)

### v0.2 — Crawling
- [ ] External template repo support
- [ ] Template caching + refresh
- [ ] 3-4 starter templates (node-api, python-cli, go-tool, static-site)
- [ ] Configuration file (~/.incubator/config.yaml)
- [ ] Confirmation screen with file preview

### v0.3 — Running
- [ ] Auto-update (self + templates)
- [ ] Conditional file inclusion in templates
- [ ] Post-create hooks
- [ ] Template search/filter in TUI
- [ ] Progress animation

### v0.4 — Polish
- [ ] `incubator config` TUI for settings
- [ ] Open in editor after create
- [ ] Custom template repos (not just the official one)
- [ ] Community template registry
- [ ] Windows support

### Future Ideas
- [ ] `incubator doctor` — check deps (git, gh, docker)
- [ ] Template versioning + changelogs
- [ ] Project presets (save a combo of template + answers for reuse)
- [ ] Team/org templates (private repos)
- [ ] Plugin system for custom providers (GitLab, Bitbucket)

---

## Repo Structure

```
sloth-incubator/
├── cmd/
│   └── incubator/
│       └── main.go
├── internal/
│   ├── tui/                # Bubbletea models
│   │   ├── app.go
│   │   ├── picker.go       # template selection
│   │   ├── form.go         # dynamic prompts
│   │   ├── confirm.go      # preview + confirm
│   │   ├── progress.go     # creation progress
│   │   └── styles.go       # Lipgloss styles
│   ├── template/           # template engine
│   │   ├── loader.go
│   │   ├── renderer.go
│   │   └── cache.go
│   ├── git/                # git + github ops
│   │   ├── init.go
│   │   └── github.go
│   ├── config/             # user config
│   │   └── config.go
│   └── updater/            # self-update
│       └── updater.go
├── install.sh              # curl one-liner target
├── .goreleaser.yaml        # cross-platform builds
├── go.mod
├── go.sum
├── PLAN.md
├── CLAUDE.md
├── README.md
└── LICENSE
```

---

## Open Questions

- [ ] Should `incubator` require `gh` CLI or handle GitHub API directly?
  - Pro gh: simpler, user probably has it, handles auth
  - Pro direct: fewer deps, more control
  - **Leaning:** require `gh` for MVP, consider direct API later

- [ ] Template inheritance? (e.g., all templates inherit a base devcontainer)
  - Nice to have but adds complexity
  - **Leaning:** defer to v0.4+

- [ ] Should the binary be called `incubator` or `sloth`?
  - `incubator` is descriptive
  - `sloth` is shorter to type
  - **Leaning:** `incubator` with `sloth` as alias?

---

*Last updated: 2026-02-23*
