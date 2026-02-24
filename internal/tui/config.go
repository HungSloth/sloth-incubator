package tui

import (
	"fmt"
	"strings"

	"github.com/HungSloth/sloth-incubator/internal/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ConfigField represents a single config field
type ConfigField struct {
	Label     string
	Key       string
	textInput textinput.Model
	// For toggle fields
	isToggle    bool
	toggleValue bool
	// For select fields
	isSelect      bool
	selectOptions []string
	selectCursor  int
}

// ConfigModel handles the configuration TUI
type ConfigModel struct {
	cfg    *config.Config
	fields []ConfigField
	cursor int
	saved  bool
	err    error
}

// NewConfigModel creates a new config editor model
func NewConfigModel(cfg *config.Config) ConfigModel {
	fields := []ConfigField{
		newTextField("GitHub User", "github_user", cfg.GitHubUser),
		newSelectField("Default Visibility", "default_visibility", cfg.DefaultVisibility, []string{"private", "public"}),
		newSelectField("Default License", "default_license", cfg.DefaultLicense, []string{"MIT", "Apache-2.0", "GPL-3.0", "none"}),
		newTextField("Project Directory", "project_dir", cfg.ProjectDir),
		newSelectField("Editor", "editor", cfg.Editor, []string{"none", "cursor", "code", "vim"}),
		newTextField("Template Repo", "template_repo", cfg.TemplateRepo),
		newToggleField("Auto Update Check", "auto_update_check", cfg.AutoUpdateCheck),
	}

	// Focus the first text input
	for i := range fields {
		if !fields[i].isToggle && !fields[i].isSelect {
			fields[i].textInput.Focus()
			break
		}
	}

	return ConfigModel{
		cfg:    cfg,
		fields: fields,
		cursor: 0,
	}
}

func newTextField(label, key, value string) ConfigField {
	ti := textinput.New()
	ti.SetValue(value)
	ti.Placeholder = label
	return ConfigField{
		Label:     label,
		Key:       key,
		textInput: ti,
	}
}

func newSelectField(label, key, current string, options []string) ConfigField {
	cursor := 0
	for i, opt := range options {
		if opt == current {
			cursor = i
			break
		}
	}
	return ConfigField{
		Label:         label,
		Key:           key,
		isSelect:      true,
		selectOptions: options,
		selectCursor:  cursor,
	}
}

func newToggleField(label, key string, value bool) ConfigField {
	return ConfigField{
		Label:       label,
		Key:         key,
		isToggle:    true,
		toggleValue: value,
	}
}

func (m ConfigModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "down":
			return m.nextField()
		case "shift+tab", "up":
			return m.prevField()
		case "left", "h":
			field := &m.fields[m.cursor]
			if field.isSelect {
				if field.selectCursor > 0 {
					field.selectCursor--
				}
			} else if field.isToggle {
				field.toggleValue = !field.toggleValue
			}
			return m, nil
		case "right", "l":
			field := &m.fields[m.cursor]
			if field.isSelect {
				if field.selectCursor < len(field.selectOptions)-1 {
					field.selectCursor++
				}
			} else if field.isToggle {
				field.toggleValue = !field.toggleValue
			}
			return m, nil
		case "enter":
			// Save config
			m.applyToConfig()
			if err := m.cfg.Save(); err != nil {
				m.err = err
			} else {
				m.saved = true
			}
			return m, tea.Quit
		case "esc":
			return m, tea.Quit
		}
	}

	// Update focused text input
	if m.cursor < len(m.fields) {
		field := &m.fields[m.cursor]
		if !field.isToggle && !field.isSelect {
			var cmd tea.Cmd
			field.textInput, cmd = field.textInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m ConfigModel) nextField() (tea.Model, tea.Cmd) {
	field := &m.fields[m.cursor]
	if !field.isToggle && !field.isSelect {
		field.textInput.Blur()
	}

	m.cursor++
	if m.cursor >= len(m.fields) {
		m.cursor = len(m.fields) - 1
	}

	field = &m.fields[m.cursor]
	if !field.isToggle && !field.isSelect {
		return m, field.textInput.Focus()
	}
	return m, nil
}

func (m ConfigModel) prevField() (tea.Model, tea.Cmd) {
	field := &m.fields[m.cursor]
	if !field.isToggle && !field.isSelect {
		field.textInput.Blur()
	}

	m.cursor--
	if m.cursor < 0 {
		m.cursor = 0
	}

	field = &m.fields[m.cursor]
	if !field.isToggle && !field.isSelect {
		return m, field.textInput.Focus()
	}
	return m, nil
}

func (m *ConfigModel) applyToConfig() {
	for _, field := range m.fields {
		switch field.Key {
		case "github_user":
			m.cfg.GitHubUser = field.textInput.Value()
		case "default_visibility":
			m.cfg.DefaultVisibility = field.selectOptions[field.selectCursor]
		case "default_license":
			m.cfg.DefaultLicense = field.selectOptions[field.selectCursor]
		case "project_dir":
			m.cfg.ProjectDir = field.textInput.Value()
		case "editor":
			m.cfg.Editor = field.selectOptions[field.selectCursor]
		case "template_repo":
			m.cfg.TemplateRepo = field.textInput.Value()
		case "auto_update_check":
			m.cfg.AutoUpdateCheck = field.toggleValue
		}
	}
}

func (m ConfigModel) View() string {
	var b strings.Builder

	header := headerStyle.Render("  Incubator Settings")
	b.WriteString(header)
	b.WriteString("\n\n")

	for i, field := range m.fields {
		isFocused := i == m.cursor
		label := field.Label + ":"
		if isFocused {
			label = focusedStyle.Render(label)
		} else {
			label = labelStyle.Render(label)
		}

		b.WriteString(fmt.Sprintf("  %-25s  ", label))

		if field.isToggle {
			if field.toggleValue {
				if isFocused {
					b.WriteString(focusedStyle.Render("[Yes]"))
					b.WriteString(mutedStyle.Render(" No "))
				} else {
					b.WriteString("[Yes]")
					b.WriteString(mutedStyle.Render(" No "))
				}
			} else {
				b.WriteString(mutedStyle.Render(" Yes "))
				if isFocused {
					b.WriteString(focusedStyle.Render("[No]"))
				} else {
					b.WriteString("[No]")
				}
			}
		} else if field.isSelect {
			for j, opt := range field.selectOptions {
				if j == field.selectCursor {
					if isFocused {
						b.WriteString(focusedStyle.Render(fmt.Sprintf("[%s]", opt)))
					} else {
						b.WriteString(fmt.Sprintf("[%s]", opt))
					}
				} else {
					b.WriteString(mutedStyle.Render(fmt.Sprintf(" %s ", opt)))
				}
				b.WriteString(" ")
			}
		} else {
			b.WriteString(field.textInput.View())
		}

		b.WriteString("\n")
	}

	if m.saved {
		b.WriteString(fmt.Sprintf("\n  %s\n", successStyle.Render("Configuration saved!")))
	}
	if m.err != nil {
		b.WriteString(fmt.Sprintf("\n  %s\n", errorStyle.Render(fmt.Sprintf("Error: %v", m.err))))
	}

	b.WriteString(helpStyle.Render("\n  tab/\u2193 next \u2022 shift+tab/\u2191 prev \u2022 \u2190/\u2192 change \u2022 enter save \u2022 esc cancel"))

	return appStyle.Render(b.String())
}
