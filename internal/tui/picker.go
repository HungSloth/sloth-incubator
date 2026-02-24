package tui

import (
	"fmt"
	"strings"

	"github.com/HungSloth/sloth-incubator/internal/template"
	tea "github.com/charmbracelet/bubbletea"
)

// PickerModel handles template selection
type PickerModel struct {
	templates []*template.TemplateManifest
	cursor    int
}

// NewPickerModel creates a new picker model
func NewPickerModel(templates []*template.TemplateManifest) PickerModel {
	return PickerModel{
		templates: templates,
		cursor:    0,
	}
}

func (m PickerModel) Init() tea.Cmd {
	return nil
}

func (m PickerModel) Update(msg tea.Msg) (PickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.templates)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.templates) > 0 {
				return m, func() tea.Msg {
					return templateSelectedMsg{manifest: m.templates[m.cursor]}
				}
			}
		case "q", "esc":
			return m, func() tea.Msg { return quitMsg{} }
		}
	}
	return m, nil
}

func (m PickerModel) View() string {
	var b strings.Builder

	// Header
	header := headerStyle.Render("  Sloth Incubator")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Template list
	for i, t := range m.templates {
		cursor := "  "
		style := inactiveItemStyle
		if i == m.cursor {
			cursor = "> "
			style = activeItemStyle
		}

		name := style.Render(t.Name)
		desc := mutedStyle.Render(t.Description)
		b.WriteString(fmt.Sprintf("%s%s  %s\n", cursor, name, desc))
	}

	// Help
	b.WriteString(helpStyle.Render("\n  ↑/↓ navigate • enter select • q quit"))

	return b.String()
}
