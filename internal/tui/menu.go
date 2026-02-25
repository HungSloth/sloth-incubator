package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type menuOption struct {
	label string
	desc  string
}

// MenuModel is the top-level action chooser.
type MenuModel struct {
	options []menuOption
	cursor  int
}

func NewMenuModel() MenuModel {
	return MenuModel{
		options: []menuOption{
			{
				label: "Create Project",
				desc:  "Pick a template and scaffold a project",
			},
			{
				label: "Create Template",
				desc:  "Launch template creator wizard",
			},
		},
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			switch m.cursor {
			case 0:
				return m, func() tea.Msg { return menuCreateProjectMsg{} }
			case 1:
				return m, func() tea.Msg { return menuCreateTemplateMsg{} }
			}
		case "q", "esc":
			return m, func() tea.Msg { return quitMsg{} }
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("  Sloth Incubator"))
	b.WriteString("\n\n")

	for i, opt := range m.options {
		cursor := "  "
		style := inactiveItemStyle
		if i == m.cursor {
			cursor = "> "
			style = activeItemStyle
		}
		b.WriteString(fmt.Sprintf("%s%s  %s\n", cursor, style.Render(opt.label), mutedStyle.Render(opt.desc)))
	}

	b.WriteString(helpStyle.Render("\n  ↑/↓ navigate • enter select • q quit"))
	return b.String()
}
