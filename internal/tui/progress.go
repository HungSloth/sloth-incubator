package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// ProgressModel handles the progress screen
type ProgressModel struct {
	manifest *template.TemplateManifest
	answers  map[string]interface{}
	steps    []ProgressStep
	current  int
	spinner  spinner.Model
	done     bool
	failed   bool

	projectDir string
	repoURL    string
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
func NewProgressModel(manifest *template.TemplateManifest, answers map[string]interface{}) ProgressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return ProgressModel{
		manifest: manifest,
		answers:  answers,
		steps: []ProgressStep{
			{Name: "Creating project directory", Status: StepPending},
			{Name: "Rendering templates", Status: StepPending},
			{Name: "Initializing git repo", Status: StepPending},
			{Name: "Creating GitHub repo", Status: StepPending},
			{Name: "Pushing to origin", Status: StepPending},
		},
		current: 0,
		spinner: s,
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
		if m.current >= 3 { // GitHub steps
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

	return func() tea.Msg {
		projectName := fmt.Sprintf("%v", answers["project_name"])
		homeDir, _ := os.UserHomeDir()
		projectDir := filepath.Join(homeDir, "projects", projectName)

		switch step {
		case 0: // Create directory
			if err := os.MkdirAll(projectDir, 0755); err != nil {
				return stepErrorMsg{err: fmt.Errorf("creating directory: %w", err)}
			}
			return stepDoneMsg{projectDir: projectDir}

		case 1: // Render templates
			renderer := template.NewRenderer(manifest, answers)
			templateFS, err := template.GetEmbeddedEmptyTemplate()
			if err != nil {
				return stepErrorMsg{err: fmt.Errorf("loading template: %w", err)}
			}
			if err := renderer.RenderTo(projectDir, templateFS); err != nil {
				return stepErrorMsg{err: fmt.Errorf("rendering templates: %w", err)}
			}
			return stepDoneMsg{}

		case 2: // Git init
			if err := git.InitRepo(projectDir); err != nil {
				return stepErrorMsg{err: err}
			}
			if err := git.InitialCommit(projectDir); err != nil {
				return stepErrorMsg{err: err}
			}
			return stepDoneMsg{}

		case 3: // Create GitHub repo
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

		case 4: // Push
			if err := git.Push(projectDir); err != nil {
				return stepErrorMsg{err: err}
			}
			return stepDoneMsg{}
		}

		return stepDoneMsg{}
	}
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

	if m.failed {
		b.WriteString(helpStyle.Render("\n  enter continue • q quit"))
	}

	return b.String()
}
