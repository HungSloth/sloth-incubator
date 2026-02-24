package tui

import (
	"fmt"
	"strings"

	"github.com/HungSloth/sloth-incubator/internal/template"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// PickerModel handles template selection with search/filter
type PickerModel struct {
	allTemplates []*template.TemplateManifest
	filtered     []*template.TemplateManifest
	cursor       int
	searchInput  textinput.Model
	searching    bool
}

// NewPickerModel creates a new picker model
func NewPickerModel(templates []*template.TemplateManifest) PickerModel {
	ti := textinput.New()
	ti.Placeholder = "Search templates..."

	return PickerModel{
		allTemplates: templates,
		filtered:     templates,
		cursor:       0,
		searchInput:  ti,
		searching:    false,
	}
}

func (m PickerModel) Init() tea.Cmd {
	return nil
}

func (m PickerModel) Update(msg tea.Msg) (PickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "esc":
				m.searching = false
				m.searchInput.Blur()
				m.searchInput.SetValue("")
				m.filtered = m.allTemplates
				m.cursor = 0
				return m, nil
			case "enter":
				if len(m.filtered) > 0 {
					return m, func() tea.Msg {
						return templateSelectedMsg{manifest: m.filtered[m.cursor]}
					}
				}
				return m, nil
			case "up":
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil
			case "down":
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.filterTemplates()
				return m, cmd
			}
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "/":
			m.searching = true
			return m, m.searchInput.Focus()
		case "enter":
			if len(m.filtered) > 0 {
				return m, func() tea.Msg {
					return templateSelectedMsg{manifest: m.filtered[m.cursor]}
				}
			}
		case "q", "esc":
			return m, func() tea.Msg { return quitMsg{} }
		}
	}
	return m, nil
}

func (m *PickerModel) filterTemplates() {
	query := strings.ToLower(m.searchInput.Value())
	if query == "" {
		m.filtered = m.allTemplates
		m.cursor = 0
		return
	}

	var filtered []*template.TemplateManifest
	for _, t := range m.allTemplates {
		name := strings.ToLower(t.Name)
		desc := strings.ToLower(t.Description)
		if strings.Contains(name, query) || strings.Contains(desc, query) {
			filtered = append(filtered, t)
		}
	}
	m.filtered = filtered
	if m.cursor >= len(m.filtered) {
		m.cursor = 0
	}
}

func (m PickerModel) View() string {
	var b strings.Builder

	// Header
	header := headerStyle.Render("  Sloth Incubator")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Search bar
	if m.searching {
		b.WriteString(fmt.Sprintf("  Search: %s\n\n", m.searchInput.View()))
	}

	// Template list
	if len(m.filtered) == 0 {
		b.WriteString(mutedStyle.Render("  No templates match your search.\n"))
	} else {
		for i, t := range m.filtered {
			cursor := "  "
			style := inactiveItemStyle
			if i == m.cursor {
				cursor = "> "
				style = activeItemStyle
			}

			name := style.Render(fmt.Sprintf("%-15s", t.Name))
			desc := mutedStyle.Render(t.Description)
			b.WriteString(fmt.Sprintf("%s%s  %s\n", cursor, name, desc))
		}
	}

	// Help
	if m.searching {
		b.WriteString(helpStyle.Render("\n  ↑/↓ navigate • enter select • esc clear search"))
	} else {
		b.WriteString(helpStyle.Render("\n  ↑/↓ navigate • enter select • / search • q quit"))
	}

	return b.String()
}
