package tui

import (
	"fmt"
	"strings"

	"github.com/HungSloth/sloth-incubator/internal/container"
	tea "github.com/charmbracelet/bubbletea"
)

// CleanModel is an interactive selector for choosing devcontainers to remove.
type CleanModel struct {
	containers []container.DevContainer
	cursor     int
	selected   map[int]bool
	cancelled  bool
}

// NewCleanModel creates a new clean selector model.
func NewCleanModel(containers []container.DevContainer) CleanModel {
	return CleanModel{
		containers: containers,
		selected:   make(map[int]bool, len(containers)),
	}
}

func (m CleanModel) Init() tea.Cmd {
	return nil
}

func (m CleanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.containers)-1 {
				m.cursor++
			}
		case " ":
			if len(m.containers) > 0 {
				m.selected[m.cursor] = !m.selected[m.cursor]
			}
		case "a":
			for i := range m.containers {
				m.selected[i] = true
			}
		case "enter":
			return m, tea.Quit
		case "esc", "q":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m CleanModel) View() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("  Clean Devcontainers"))
	b.WriteString("\n\n")

	if len(m.containers) == 0 {
		b.WriteString(mutedStyle.Render("  No devcontainers found.\n"))
		b.WriteString(helpStyle.Render("\n  q quit"))
		return b.String()
	}

	for i, c := range m.containers {
		cursor := "  "
		style := inactiveItemStyle
		if i == m.cursor {
			cursor = "> "
			style = activeItemStyle
		}

		check := "[ ]"
		if m.selected[i] {
			check = "[x]"
		}

		status := c.Status
		if strings.HasPrefix(status, "Up ") || strings.HasPrefix(status, "Restarting") {
			status = successStyle.Render(status)
		} else {
			status = mutedStyle.Render(status)
		}

		line := fmt.Sprintf("%s%s %s  %s  %s", cursor, check, style.Render(c.Name), status, mutedStyle.Render(c.ProjectDir))
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("\n  ↑/↓ navigate • space toggle • a select all • enter confirm • q cancel"))
	return b.String()
}

// SelectedContainers returns the currently selected containers.
func (m CleanModel) SelectedContainers() []container.DevContainer {
	selected := make([]container.DevContainer, 0, len(m.selected))
	for i, c := range m.containers {
		if m.selected[i] {
			selected = append(selected, c)
		}
	}
	return selected
}

// Cancelled reports whether the user exited without confirming.
func (m CleanModel) Cancelled() bool {
	return m.cancelled
}

// RunCleanSelection runs the interactive selector and returns the chosen containers.
func RunCleanSelection(containers []container.DevContainer) ([]container.DevContainer, bool, error) {
	model := NewCleanModel(containers)
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, false, err
	}

	cleanModel, ok := finalModel.(CleanModel)
	if !ok {
		return nil, false, fmt.Errorf("unexpected clean model type")
	}

	return cleanModel.SelectedContainers(), cleanModel.Cancelled(), nil
}
