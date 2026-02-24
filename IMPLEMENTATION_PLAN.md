# Sloth Incubator — Implementation Plan

This document breaks down the full implementation of Sloth Incubator into ordered, actionable steps across all milestones. Each step includes the files to create/modify, key implementation details, and dependencies.

---

## v0.1 — Walking (MVP)

**Goal:** A working end-to-end flow: launch TUI → pick the single built-in template → answer prompts → get a scaffolded project with git + GitHub remote.

### Step 1: Go Project Scaffold

**Files to create:**
- `go.mod` — module `github.com/HungSloth/sloth-incubator`, Go 1.22+
- `cmd/incubator/main.go` — entrypoint, wire up CLI commands
- `CLAUDE.md` — development instructions for AI assistants

**Details:**
- Initialize with `go mod init github.com/HungSloth/sloth-incubator`
- `main.go` uses `cobra` for the CLI framework (or keep it minimal with just `os.Args` for MVP)
- Default command (no args) launches the TUI
- Subcommands: `new`, `list`, `version` (stubs for MVP)

**Dependencies:** None — this is the foundation.

---

### Step 2: TUI Framework + App Shell

**Files to create:**
- `internal/tui/app.go` — root Bubbletea model, manages screen transitions
- `internal/tui/styles.go` — shared Lipgloss styles (header, borders, colors, etc.)

**Key dependencies to add:**
```
github.com/charmbracelet/bubbletea
github.com/charmbracelet/lipgloss
github.com/charmbracelet/bubbles
```

**Details:**
- `app.go` defines a top-level `Model` that holds the current screen state (enum: `screenPicker`, `screenForm`, `screenConfirm`, `screenProgress`, `screenDone`)
- The model delegates `Update()` and `View()` to the active screen's sub-model
- Use alternate screen buffer (`tea.EnterAltScreen`) for clean entry/exit
- `styles.go` defines reusable styles: header box, active/inactive list items, muted text, accent colors

**Dependencies:** Step 1

---

### Step 3: Template Selection Screen (Picker)

**Files to create:**
- `internal/tui/picker.go` — template selection list using `bubbles/list`

**Details:**
- Uses `bubbles/list` component with custom item delegate for the template display format
- For MVP, this is a hardcoded list with a single item: "empty" (blank project with devcontainer)
- On `enter`, transitions to the form screen, passing the selected template
- Displays the sloth incubator header banner

**Dependencies:** Step 2

---

### Step 4: Dynamic Form Screen

**Files to create:**
- `internal/tui/form.go` — dynamic prompts driven by template manifest
- `internal/template/manifest.go` — types for `template.yaml` (Prompt, TemplateManifest, etc.)

**Details:**
- Form screen renders prompts defined in `TemplateManifest.Prompts`
- Supported prompt types for MVP: `text`, `select`, `confirm`
- Use `bubbles/textinput` for text fields, custom list for selects, toggle for confirms
- Tab/Shift+Tab navigates between fields
- Enter on last field (or a "Continue" button) advances to confirmation
- Collects all answers into a `map[string]any` for template rendering
- For the MVP "empty" template, prompts are: project name, description, visibility, license

**Dependencies:** Step 3

---

### Step 5: Confirmation Screen

**Files to create:**
- `internal/tui/confirm.go` — summary view before project creation

**Details:**
- Displays a read-only summary: project name, template, visibility, license, and list of files to create
- `enter` proceeds to creation (progress screen)
- `esc` goes back to form
- `q` cancels entirely
- The file list is derived from the template (hardcoded for MVP)

**Dependencies:** Step 4

---

### Step 6: Template Engine

**Files to create:**
- `internal/template/renderer.go` — renders Go templates with user answers
- `internal/template/embedded.go` — embedded template files for the MVP "empty" template

**Details:**
- Use Go `embed` to bundle the built-in "empty" template directly in the binary
- Template files: `README.md.tmpl`, `.gitignore`, `.devcontainer/devcontainer.json.tmpl`, `CLAUDE.md.tmpl`
- `renderer.go` walks the template directory, processes `.tmpl` files through `text/template`, copies non-tmpl files as-is
- Template variables come from the form's `map[string]any`
- Directory names with `{{name}}` are expanded
- Output goes to the target project directory (from config or `~/projects/`)

**Embedded "empty" template contents:**
- `README.md.tmpl` — project name + description
- `.gitignore` — sensible defaults
- `.devcontainer/devcontainer.json.tmpl` — basic Ubuntu devcontainer with git + gh features
- `CLAUDE.md.tmpl` — basic project instructions

**Dependencies:** Step 4 (for manifest types)

---

### Step 7: Git + GitHub Integration

**Files to create:**
- `internal/git/init.go` — `git init`, initial commit
- `internal/git/github.go` — create GitHub repo via `gh`, set remote, push

**Details:**
- `init.go`: Run `git init` in the project directory, `git add .`, `git commit -m "Initial commit from sloth-incubator"`
- `github.go`: Shell out to `gh repo create <name> --private/--public --source=. --remote=origin --push`
- Check that `git` and `gh` are available on PATH before starting; surface clear error if not
- Functions return errors that the progress screen can display

**Dependencies:** Step 6

---

### Step 8: Progress + Done Screens

**Files to create:**
- `internal/tui/progress.go` — step-by-step creation progress
- `internal/tui/done.go` — completion summary with next steps

**Details:**
- Progress screen shows a checklist of steps: create directory → render templates → git init → create GitHub repo → push
- Each step updates status (pending → in progress → done/failed) as a `tea.Cmd` completes
- Uses a spinner (`bubbles/spinner`) on the active step
- If a step fails, show the error and offer to retry or quit
- Done screen shows local path, GitHub URL, and the `cd` command
- `enter` on done screen opens the project in the configured editor (MVP: just print the path)
- `q` exits cleanly

**Dependencies:** Steps 6, 7

---

### Step 9: CLI Wiring + Install Script

**Files to modify:**
- `cmd/incubator/main.go` — wire TUI launch to default command and `new` subcommand

**Files to create:**
- `install.sh` — curl one-liner install script
- `.goreleaser.yaml` — cross-compilation config

**Details:**
- `main.go`: default (no subcommand) and `new` both launch the TUI via `tea.NewProgram(tui.NewApp())`
- `version` prints the version (set via `-ldflags` at build time)
- `list` prints available templates (just "empty" for MVP)
- `install.sh`: detect OS/arch, download from GitHub Releases, place in `~/.local/bin`, verify
- `.goreleaser.yaml`: build for darwin/arm64, darwin/amd64, linux/arm64, linux/amd64

**Dependencies:** Steps 2-8

---

### v0.1 Deliverable Checklist
- [ ] `incubator` launches TUI
- [ ] Single "empty" template available
- [ ] User fills in project name, description, visibility, license
- [ ] Confirmation screen shows summary
- [ ] Project directory created with rendered templates
- [ ] Git repo initialized with initial commit
- [ ] GitHub repo created and code pushed
- [ ] `install.sh` works on macOS and Linux
- [ ] `incubator version` prints version

---

## v0.2 — Crawling

**Goal:** Support external templates from a separate repo, add caching, ship 3-4 real templates, and add user configuration.

### Step 10: Template Loader (Remote Repo)

**Files to create:**
- `internal/template/loader.go` — fetch templates from the remote `incubator-templates` repo
- `internal/template/registry.go` — parse `registry.yaml` index

**Details:**
- Clone/pull `incubator-templates` repo (shallow clone for speed) into `~/.incubator/templates/`
- Parse `registry.yaml` to get the list of available templates with metadata
- Each template directory contains a `template.yaml` manifest
- Loader returns `[]TemplateManifest` for the picker screen
- Falls back to embedded templates if fetch fails (offline mode)

**Dependencies:** v0.1 complete

---

### Step 11: Template Caching

**Files to create:**
- `internal/template/cache.go` — cache management with TTL

**Details:**
- Store last-fetch timestamp in `~/.incubator/cache.json`
- On launch: if last fetch was >24 hours ago, trigger background refresh (non-blocking)
- `incubator update` forces a fresh pull
- Cache stores the shallow git clone; updates are just `git pull`
- If cache directory doesn't exist, do a blocking fetch on first run

**Dependencies:** Step 10

---

### Step 12: Configuration System

**Files to create:**
- `internal/config/config.go` — load/save `~/.incubator/config.yaml`

**Details:**
- Config struct: `GitHubUser`, `DefaultVisibility`, `DefaultLicense`, `ProjectDir`, `Editor`, `TemplateRepo`, `AutoUpdateCheck`
- Load on startup; create with defaults if missing
- `incubator config` subcommand prints current config (no TUI editor yet — that's v0.4)
- Config values are used as defaults in the form screen (user can override per-project)
- Populate `GitHubUser` by running `gh api user -q .login` on first run if not set

**Dependencies:** v0.1 complete

---

### Step 13: Starter Templates

**Repo:** `incubator-templates` (separate repository)

**Templates to create:**
1. **node-api** — Node.js API with Express/Fastify/Hono choice, optional Postgres/SQLite, Dockerfile, CI
2. **python-cli** — Python CLI with Click/Typer, pyproject.toml, virtual env setup
3. **go-tool** — Go CLI with Cobra, goreleaser, CI
4. **static-site** — HTML/CSS/JS with optional Tailwind, simple dev server

Each template includes:
- `template.yaml` manifest with prompts
- Template files (`.tmpl` for dynamic, plain for static)
- `.devcontainer/devcontainer.json.tmpl` with appropriate features
- `CLAUDE.md.tmpl` with project-specific instructions
- `.gitignore` appropriate for the language

**Dependencies:** Steps 10, 11

---

### Step 14: Enhanced Confirmation Screen

**Files to modify:**
- `internal/tui/confirm.go` — show actual file tree from template

**Details:**
- Instead of hardcoded file list, walk the template directory and show what will be rendered
- Resolve conditional files based on user answers
- Show file count and approximate project size
- Visually distinguish template-generated files from static files

**Dependencies:** Step 10

---

### Step 15: Update Command

**Files to modify:**
- `cmd/incubator/main.go` — wire `update` subcommand

**Details:**
- `incubator update` triggers: template cache refresh + check for binary update
- For now, binary update just prints "check GitHub Releases for latest version"
- Template refresh does a `git pull` on the cached templates repo
- Show what changed (new/updated templates)

**Dependencies:** Steps 11, 12

---

### v0.2 Deliverable Checklist
- [ ] Templates load from `incubator-templates` repo
- [ ] Template caching with 24h TTL
- [ ] 4 starter templates working end-to-end
- [ ] `~/.incubator/config.yaml` with user settings
- [ ] Config values populate form defaults
- [ ] Confirmation screen shows actual file tree
- [ ] `incubator update` refreshes templates
- [ ] `incubator list` shows all available templates

---

## v0.3 — Running

**Goal:** Auto-update, conditional template logic, post-create hooks, search/filter, and polished progress UI.

### Step 16: Self-Update

**Files to create:**
- `internal/updater/updater.go` — check GitHub Releases, download + replace binary

**Details:**
- Background goroutine on launch: `GET /repos/HungSloth/sloth-incubator/releases/latest`
- Compare current version (embedded at build time) with latest release tag
- If newer: show subtle banner at bottom of TUI: `"Update available: v1.2.0 → v1.3.0 (run incubator update)"`
- `incubator update`: download the asset matching current OS/arch, verify checksum if available, replace the running binary (`os.Rename` dance), print success
- Never block the user; update check is fire-and-forget

**Dependencies:** v0.2 complete

---

### Step 17: Conditional File Inclusion

**Files to modify:**
- `internal/template/renderer.go` — evaluate `when` conditions in template manifests

**Details:**
- Template `files` section in `template.yaml` has `when` expressions
- Evaluate `when` using Go templates: if the expression renders to a truthy value, include the file group
- Support: `{{ .docker }}` (boolean), `{{ ne .database "none" }}` (comparison)
- Walk the template, filter files based on conditions before rendering
- Handle glob patterns in `src` field (e.g., `docker/**`)

**Dependencies:** v0.2 complete

---

### Step 18: Post-Create Hooks

**Files to create:**
- `internal/template/hooks.go` — execute post-create scripts

**Details:**
- After template rendering and git setup, check for `hooks.post_create` in manifest
- Execute the hook script in the project directory with `os/exec`
- Pipe stdout/stderr to the progress screen
- Hook failure is non-fatal: warn the user but don't roll back
- Hooks run with template variables available as environment variables (`INCUBATOR_PROJECT_NAME`, etc.)

**Dependencies:** Step 17

---

### Step 19: Template Search/Filter

**Files to modify:**
- `internal/tui/picker.go` — add fuzzy search to template list

**Details:**
- Add a text input at the top of the picker for search
- Filter templates by name, description, and tags as the user types
- Use `bubbles/textinput` for the search field
- `/` key focuses the search input (vim-style)
- `esc` clears search and returns focus to list

**Dependencies:** v0.2 complete

---

### Step 20: Progress Animation Polish

**Files to modify:**
- `internal/tui/progress.go` — animated progress bar + better step display

**Details:**
- Add a `bubbles/progress` bar that fills as steps complete
- Each step has an estimated weight (e.g., render templates: 20%, git init: 10%, GitHub create: 40%, push: 30%)
- Smooth animation between step completions
- Show elapsed time
- Better error display with full error message in a scrollable viewport

**Dependencies:** v0.2 complete

---

### v0.3 Deliverable Checklist
- [ ] Auto-update check on launch with non-blocking banner
- [ ] `incubator update` self-updates the binary
- [ ] Conditional file inclusion working (docker, CI, database files)
- [ ] Post-create hooks execute after project creation
- [ ] Template search/filter in picker
- [ ] Animated progress bar with step weights

---

## v0.4 — Polish

**Goal:** Configuration TUI, editor integration, custom template repos, and Windows support.

### Step 21: Configuration TUI

**Files to create:**
- `internal/tui/config.go` — interactive settings editor

**Details:**
- `incubator config` launches a TUI form for editing settings
- Fields for: GitHub user, default visibility, default license, project directory, editor, template repo, auto-update toggle
- Save on confirm; cancel discards changes
- Validate paths (project directory exists or offer to create)
- Reuse the form components from Step 4

**Dependencies:** v0.3 complete

---

### Step 22: Open in Editor After Create

**Files to modify:**
- `internal/tui/done.go` — `enter` opens project in configured editor

**Details:**
- Read `editor` from config (cursor, code, vim, none)
- Map editor names to commands: `cursor` → `cursor`, `code` → `code`, `vim` → `vim`
- On `enter` at done screen: `exec.Command(editor, projectPath).Start()`
- For `code` and `cursor`: pass `--new-window` flag
- If editor is `none` or not found, just print the path

**Dependencies:** Step 21

---

### Step 23: Custom Template Repos

**Files to modify:**
- `internal/template/loader.go` — support multiple template sources
- `internal/config/config.go` — `template_repos` list instead of single `template_repo`

**Details:**
- Config changes from `template_repo: string` to `template_repos: []string` (keep backward compat)
- Loader fetches from all configured repos and merges template lists
- Templates namespaced by repo: `official/node-api`, `myorg/custom-template`
- Picker shows repo source as a subtle tag
- Conflict resolution: if same template name in multiple repos, show all with source labels
- `incubator config` lets users add/remove template repos

**Dependencies:** Step 21

---

### Step 24: Community Template Registry

**Files to create:**
- `internal/template/registry_remote.go` — fetch community template index

**Details:**
- A central `registry.yaml` hosted in the main `sloth-incubator` repo (or a dedicated registry repo)
- Lists community template repos with metadata (author, description, stars, verified status)
- `incubator list --community` shows community templates
- `incubator add-repo <url>` adds a community repo to local config
- Basic trust model: verified templates from known orgs get a badge

**Dependencies:** Step 23

---

### Step 25: Windows Support

**Files to modify:**
- Various — path handling, shell commands, install script

**Details:**
- Replace `/` path separators with `filepath.Join` everywhere (audit all path construction)
- Windows-specific `gh` and `git` command handling (no `sh -c` wrappers)
- PowerShell install script: `install.ps1`
- `.goreleaser.yaml`: add `windows/amd64` target
- Test devcontainer creation on Windows (Docker Desktop)
- Handle `%USERPROFILE%` vs `$HOME` for config directory (`os.UserHomeDir()` handles this)

**Dependencies:** v0.3 complete

---

### v0.4 Deliverable Checklist
- [ ] `incubator config` launches interactive settings TUI
- [ ] `enter` on done screen opens project in editor
- [ ] Multiple template repos supported
- [ ] Community template registry browsable
- [ ] Windows amd64 binary builds and works
- [ ] PowerShell install script

---

## Implementation Order Summary

```
v0.1 MVP (Steps 1-9)
│
├── Step 1: Go project scaffold
├── Step 2: TUI framework + app shell
├── Step 3: Template picker screen
├── Step 4: Dynamic form screen
├── Step 5: Confirmation screen
├── Step 6: Template engine + embedded template
├── Step 7: Git + GitHub integration
├── Step 8: Progress + done screens
└── Step 9: CLI wiring + install script
│
v0.2 Crawling (Steps 10-15)
│
├── Step 10: Template loader (remote repo)
├── Step 11: Template caching
├── Step 12: Configuration system
├── Step 13: Starter templates (separate repo)
├── Step 14: Enhanced confirmation screen
└── Step 15: Update command
│
v0.3 Running (Steps 16-20)
│
├── Step 16: Self-update
├── Step 17: Conditional file inclusion
├── Step 18: Post-create hooks
├── Step 19: Template search/filter
└── Step 20: Progress animation polish
│
v0.4 Polish (Steps 21-25)
│
├── Step 21: Configuration TUI
├── Step 22: Open in editor
├── Step 23: Custom template repos
├── Step 24: Community template registry
└── Step 25: Windows support
```

---

## Key Technical Decisions

### Package Management
- Use Go modules. Pin all dependencies to specific versions.
- Key deps: `bubbletea`, `lipgloss`, `bubbles`, `cobra` (CLI), `gopkg.in/yaml.v3` (YAML parsing).

### Error Handling Strategy
- Functions return `error`. No panics.
- TUI displays errors inline (red text in the progress screen).
- Git/GitHub errors suggest fixes: "Is `gh` installed?", "Run `gh auth login`", etc.

### Testing Strategy
- Unit tests for `internal/template/` (renderer, cache, manifest parsing)
- Unit tests for `internal/config/` (load, save, defaults)
- Integration tests for `internal/git/` (use temp dirs, mock `gh` with a script)
- TUI tests are hard to unit test — rely on manual testing and screenshot tests if needed
- CI runs `go test ./...` on every push

### Build + Release
- GoReleaser for cross-compilation and GitHub Releases
- Version set via `-ldflags "-X main.version=v0.1.0"` at build time
- GitHub Actions workflow: test on push, release on tag

---

## Open Decisions to Resolve Before Starting

1. **CLI framework:** Use `cobra` (heavier, more features) or plain `os.Args` + flag (lighter)? Recommendation: `cobra` — it's the Go standard, handles subcommands well, and adds help text for free.

2. **Template repo bootstrap:** Should `incubator` embed a fallback copy of the templates repo, or require network on first run? Recommendation: embed the "empty" template only; all others require a fetch.

3. **Binary name:** `incubator` vs `sloth`? Recommendation: build as `incubator`, decide on alias later (can always add a symlink in the install script).
