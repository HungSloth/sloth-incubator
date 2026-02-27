package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/HungSloth/sloth-incubator/internal/config"
	"github.com/HungSloth/sloth-incubator/internal/git"
	"github.com/HungSloth/sloth-incubator/internal/template"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// StepStatus represents the status of a progress step
type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepDone
	StepFailed
)

// ProgressStep represents a single step in the creation process
type ProgressStep struct {
	Name   string
	Status StepStatus
	Error  string
}

const (
	stepCreateProjectDir = "Creating project directory"
	stepRenderTemplates  = "Rendering templates"
	stepInitGitRepo      = "Initializing git repo"
	stepCommitScaffold   = "Committing scaffold files"
	stepCreateGitHubRepo = "Creating GitHub repo"
	stepPushToOrigin     = "Pushing to origin"
)

// ProgressModel handles the progress screen
type ProgressModel struct {
	manifest *template.TemplateManifest
	answers  map[string]interface{}
	cfg      *config.Config
	steps    []ProgressStep
	current  int
	spinner  spinner.Model
	done     bool
	failed   bool

	projectDir string
	repoURL    string
	initMode   bool
	initDir    string
}

// Step result messages
type stepDoneMsg struct {
	projectDir string
	repoURL    string
}

type stepErrorMsg struct {
	err error
}

// NewProgressModel creates a new progress model
func NewProgressModel(manifest *template.TemplateManifest, answers map[string]interface{}, cfg *config.Config, initMode bool, initDir string) ProgressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot

	steps := []ProgressStep{
		{Name: stepRenderTemplates, Status: StepPending},
	}
	if initMode {
		steps = append(steps, ProgressStep{Name: stepCommitScaffold, Status: StepPending})
	} else {
		steps = append([]ProgressStep{{Name: stepCreateProjectDir, Status: StepPending}}, steps...)
		steps = append(steps, ProgressStep{Name: stepInitGitRepo, Status: StepPending})
	}
	if !initMode && shouldCreateGitHubRepo(answers) {
		steps = append(steps,
			ProgressStep{Name: stepCreateGitHubRepo, Status: StepPending},
			ProgressStep{Name: stepPushToOrigin, Status: StepPending},
		)
	}

	return ProgressModel{
		manifest: manifest,
		answers:  answers,
		cfg:      cfg,
		steps:    steps,
		current: 0,
		spinner: s,
		initMode: initMode,
		initDir:  initDir,
	}
}

func (m ProgressModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runCurrentStep())
}

func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case stepDoneMsg:
		m.steps[m.current].Status = StepDone
		if msg.projectDir != "" {
			m.projectDir = msg.projectDir
		}
		if msg.repoURL != "" {
			m.repoURL = msg.repoURL
		}

		m.current++
		if m.current >= len(m.steps) {
			m.done = true
			return m, func() tea.Msg {
				return progressDoneMsg{
					projectDir: m.projectDir,
					repoURL:    m.repoURL,
				}
			}
		}

		m.steps[m.current].Status = StepRunning
		return m, m.runCurrentStep()

	case stepErrorMsg:
		m.steps[m.current].Status = StepFailed
		m.steps[m.current].Error = msg.err.Error()
		m.failed = true

		// If git/github steps fail, still try to finish with what we have
		if isGitHubStep(m.steps[m.current].Name) {
			m.current++
			if m.current >= len(m.steps) {
				m.done = true
				return m, func() tea.Msg {
					return progressDoneMsg{
						projectDir: m.projectDir,
						repoURL:    m.repoURL,
					}
				}
			}
			m.steps[m.current].Status = StepRunning
			return m, m.runCurrentStep()
		}
		return m, nil

	case tea.KeyMsg:
		if m.failed || m.done {
			switch msg.String() {
			case "q", "esc":
				return m, func() tea.Msg { return quitMsg{} }
			case "enter":
				if m.done || m.projectDir != "" {
					return m, func() tea.Msg {
						return progressDoneMsg{
							projectDir: m.projectDir,
							repoURL:    m.repoURL,
						}
					}
				}
			}
		}
	}

	return m, nil
}

func (m ProgressModel) runCurrentStep() tea.Cmd {
	step := m.current
	answers := m.answers
	manifest := m.manifest
	cfg := m.cfg
	initMode := m.initMode
	initDir := m.initDir

	return func() tea.Msg {
		projectDir := initDir
		if !initMode {
			projectName := fmt.Sprintf("%v", answers["project_name"])
			baseDir := filepath.Join(os.Getenv("HOME"), "projects")
			if cfg != nil {
				baseDir = cfg.GetProjectDir()
			}
			projectDir = filepath.Join(baseDir, projectName)
		}
		stepName := m.steps[step].Name

		switch stepName {
		case stepCreateProjectDir:
			if err := os.MkdirAll(projectDir, 0755); err != nil {
				return stepErrorMsg{err: fmt.Errorf("creating directory: %w", err)}
			}
			return stepDoneMsg{projectDir: projectDir}

		case stepRenderTemplates:
			renderer := template.NewRenderer(manifest, answers)
			renderer.SkipExisting = initMode
			templateRepo := config.DefaultConfig().TemplateRepo
			if cfg != nil && cfg.TemplateRepo != "" {
				templateRepo = cfg.TemplateRepo
			}
			templateFS, err := template.ResolveTemplateFS(manifest, config.ConfigDir(), templateRepo)
			if err != nil {
				return stepErrorMsg{err: fmt.Errorf("loading template: %w", err)}
			}
			if err := renderer.RenderTo(projectDir, templateFS); err != nil {
				return stepErrorMsg{err: fmt.Errorf("rendering templates: %w", err)}
			}
			return stepDoneMsg{}

		case stepInitGitRepo:
			if err := git.InitRepo(projectDir); err != nil {
				return stepErrorMsg{err: err}
			}
			if err := git.InitialCommit(projectDir); err != nil {
				return stepErrorMsg{err: err}
			}
			return stepDoneMsg{}

		case stepCommitScaffold:
			if git.HasRepo(projectDir) {
				if err := git.CommitAll(projectDir, "Add incubator scaffolding"); err != nil {
					if strings.Contains(err.Error(), "nothing to commit") {
						return stepDoneMsg{}
					}
					return stepErrorMsg{err: err}
				}
				return stepDoneMsg{}
			}
			if err := git.InitRepo(projectDir); err != nil {
				return stepErrorMsg{err: err}
			}
			if err := git.InitialCommit(projectDir); err != nil {
				return stepErrorMsg{err: err}
			}
			return stepDoneMsg{}

		case stepCreateGitHubRepo:
			if !shouldCreateGitHubRepo(answers) {
				return stepDoneMsg{}
			}

			projectName := fmt.Sprintf("%v", answers["project_name"])
			visibility := "private"
			if v, ok := answers["visibility"]; ok {
				visibility = fmt.Sprintf("%v", v)
			}
			isPrivate := visibility == "private"

			repoURL, err := git.CreateRepo(projectName, isPrivate, projectDir)
			if err != nil {
				return stepErrorMsg{err: err}
			}
			return stepDoneMsg{repoURL: repoURL}

		case stepPushToOrigin:
			if !shouldCreateGitHubRepo(answers) {
				return stepDoneMsg{}
			}

			if err := git.Push(projectDir); err != nil {
				return stepErrorMsg{err: err}
			}
			return stepDoneMsg{}
		}

		return stepDoneMsg{}
	}
}

// progressPercent returns the overall progress percentage
func (m ProgressModel) progressPercent() float64 {
	weights := map[string]float64{
		stepCreateProjectDir: 0.10,
		stepRenderTemplates:  0.30,
		stepInitGitRepo:      0.10,
		stepCommitScaffold:   0.20,
		stepCreateGitHubRepo: 0.40,
		stepPushToOrigin:     0.20,
	}
	var totalWeight float64
	for _, step := range m.steps {
		totalWeight += weights[step.Name]
	}
	if totalWeight == 0 {
		return 0
	}

	var total float64
	for _, step := range m.steps {
		weight := weights[step.Name] / totalWeight
		switch step.Status {
		case StepDone:
			total += weight
		case StepRunning:
			total += weight * 0.5 // half credit for running
		case StepFailed:
			total += weight // count as done for progress purposes
		}
	}
	return total
}

func shouldCreateGitHubRepo(answers map[string]interface{}) bool {
	if answers == nil {
		return true
	}
	value, ok := answers["create_github_repo"]
	if !ok {
		return true
	}
	enabled, ok := value.(bool)
	if !ok {
		return true
	}
	return enabled
}

func isGitHubStep(stepName string) bool {
	return stepName == stepCreateGitHubRepo || stepName == stepPushToOrigin
}

// renderProgressBar renders a visual progress bar
func renderProgressBar(percent float64, width int) string {
	if width < 4 {
		width = 20
	}
	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("[%s] %d%%", bar, int(percent*100))
}

func (m ProgressModel) View() string {
	var b strings.Builder

	b.WriteString("\n")

	for i, step := range m.steps {
		var icon string
		switch step.Status {
		case StepPending:
			icon = mutedStyle.Render("○")
		case StepRunning:
			icon = m.spinner.View()
		case StepDone:
			icon = successStyle.Render("✓")
		case StepFailed:
			icon = errorStyle.Render("✗")
		}

		name := step.Name
		if i == m.current && step.Status == StepRunning {
			name = name + "..."
		}

		b.WriteString(fmt.Sprintf("  %s %s\n", icon, name))

		if step.Status == StepFailed && step.Error != "" {
			b.WriteString(fmt.Sprintf("    %s\n", errorStyle.Render(step.Error)))
		}
	}

	// Progress bar
	b.WriteString("\n")
	percent := m.progressPercent()
	bar := renderProgressBar(percent, 30)
	if percent >= 1.0 {
		b.WriteString(fmt.Sprintf("  %s\n", successStyle.Render(bar)))
	} else {
		b.WriteString(fmt.Sprintf("  %s\n", focusedStyle.Render(bar)))
	}

	if m.failed {
		b.WriteString(helpStyle.Render("\n  enter continue • q quit"))
	}

	return b.String()
}
