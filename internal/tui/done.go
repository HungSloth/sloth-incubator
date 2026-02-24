package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// DoneModel handles the completion screen
type DoneModel struct {
	projectDir string
	repoURL    string
}

// NewDoneModel creates a new done model
func NewDoneModel(projectDir, repoURL string) DoneModel {
	return DoneModel{
		projectDir: projectDir,
		repoURL:    repoURL,
	}
}

func (m DoneModel) Init() tea.Cmd {
	return nil
}

func (m DoneModel) Update(msg tea.Msg) (DoneModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "enter":
			return m, func() tea.Msg { return quitMsg{} }
		}
	}
	return m, nil
}

func (m DoneModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(successStyle.Render("  Project is ready!"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Render("Local:"), valueStyle.Render(m.projectDir)))

	if m.repoURL != "" {
		b.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Render("Remote:"), valueStyle.Render(m.repoURL)))
	}

	b.WriteString(fmt.Sprintf("\n  %s\n", mutedStyle.Render(fmt.Sprintf("cd %s", m.projectDir))))

	b.WriteString(helpStyle.Render("\n  enter/q exit"))

	return b.String()
}
