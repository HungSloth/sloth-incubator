package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/HungSloth/sloth-incubator/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

// DoneModel handles the completion screen
type DoneModel struct {
	projectDir string
	repoURL    string
	editor     string
	initMode   bool
}

// NewDoneModel creates a new done model
func NewDoneModel(projectDir, repoURL string, initMode bool) DoneModel {
	editor := "none"
	if cfg, err := config.Load(); err == nil {
		editor = cfg.Editor
	}

	return DoneModel{
		projectDir: projectDir,
		repoURL:    repoURL,
		editor:     editor,
		initMode:   initMode,
	}
}

func (m DoneModel) Init() tea.Cmd {
	return nil
}

// openInEditorMsg is sent after attempting to open the project in an editor
type openInEditorMsg struct {
	err error
}

func (m DoneModel) Update(msg tea.Msg) (DoneModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.editor != "" && m.editor != "none" {
				return m, m.openInEditor()
			}
			return m, func() tea.Msg { return quitMsg{} }
		case "q", "esc":
			return m, func() tea.Msg { return quitMsg{} }
		}
	case openInEditorMsg:
		return m, func() tea.Msg { return quitMsg{} }
	}
	return m, nil
}

func (m DoneModel) openInEditor() tea.Cmd {
	return func() tea.Msg {
		editorCmd := m.editor
		var args []string

		switch m.editor {
		case "cursor":
			editorCmd = "cursor"
			args = []string{"--new-window", m.projectDir}
		case "code":
			editorCmd = "code"
			args = []string{"--new-window", m.projectDir}
		case "vim":
			editorCmd = "vim"
			args = []string{m.projectDir}
		default:
			args = []string{m.projectDir}
		}

		cmd := exec.Command(editorCmd, args...)
		err := cmd.Start()
		return openInEditorMsg{err: err}
	}
}

func (m DoneModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	header := "  Project is ready!"
	if m.initMode {
		header = "  Project initialized!"
	}
	b.WriteString(successStyle.Render(header))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Render("Local:"), valueStyle.Render(m.projectDir)))

	if m.repoURL != "" {
		b.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Render("Remote:"), valueStyle.Render(m.repoURL)))
	}

	b.WriteString(fmt.Sprintf("\n  %s\n", mutedStyle.Render(fmt.Sprintf("cd %s", m.projectDir))))

	if m.editor != "" && m.editor != "none" {
		b.WriteString(helpStyle.Render(fmt.Sprintf("\n  enter open in %s • q exit", m.editor)))
	} else {
		b.WriteString(helpStyle.Render("\n  enter/q exit"))
	}

	return b.String()
}
