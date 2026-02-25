package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/HungSloth/sloth-incubator/internal/template"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// FormField represents a single form field (text, select, or confirm)
type FormField struct {
	prompt    template.Prompt
	textInput textinput.Model
	// For select fields
	selectOptions []template.PromptOption
	selectCursor  int
	// For confirm fields
	confirmValue bool
}

// FormModel handles the dynamic form
type FormModel struct {
	manifest *template.TemplateManifest
	fields   []FormField
	cursor   int // which field is focused
}

// NewFormModel creates a new form from a template manifest
func NewFormModel(manifest *template.TemplateManifest) FormModel {
	fields := make([]FormField, len(manifest.Prompts))

	for i, p := range manifest.Prompts {
		field := FormField{
			prompt: p,
		}

		switch p.Type {
		case template.PromptText:
			ti := textinput.New()
			ti.Placeholder = p.Label
			if p.Default != nil {
				ti.SetValue(fmt.Sprintf("%v", p.Default))
			}
			if i == 0 {
				ti.Focus()
			}
			field.textInput = ti

		case template.PromptSelect:
			field.selectOptions = p.Options
			// Set default selection
			if p.Default != nil {
				defaultVal := fmt.Sprintf("%v", p.Default)
				for j, opt := range p.Options {
					if opt.Value == defaultVal {
						field.selectCursor = j
						break
					}
				}
			}

		case template.PromptConfirm:
			if p.Default != nil {
				if bVal, ok := p.Default.(bool); ok {
					field.confirmValue = bVal
				}
			}
		}

		fields[i] = field
	}

	return FormModel{
		manifest: manifest,
		fields:   fields,
		cursor:   0,
	}
}

func (m FormModel) Init() tea.Cmd {
	if len(m.fields) > 0 && m.fields[0].prompt.Type == template.PromptText {
		return m.fields[0].textInput.Focus()
	}
	return nil
}

func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		fieldType := "unknown"
		if m.cursor >= 0 && m.cursor < len(m.fields) {
			fieldType = string(m.fields[m.cursor].prompt.Type)
		}
		// #region agent log
		writeDebugLog("repro-1", "H1", "internal/tui/form.go:Update:90", "form received key", map[string]interface{}{
			"key":       msg.String(),
			"cursor":    m.cursor,
			"fieldType": fieldType,
		})
		// #endregion
		switch msg.String() {
		case "tab", "down":
			return m.nextField()
		case "shift+tab", "up":
			return m.prevField()
		case "enter":
			// If on the last field, submit
			if m.cursor == len(m.fields)-1 {
				return m, func() tea.Msg {
					return formCompletedMsg{answers: m.collectAnswers()}
				}
			}
			return m.nextField()
		case "esc":
			return m, func() tea.Msg { return formBackMsg{} }
		case "left", "h":
			// #region agent log
			writeDebugLog("repro-1", "H1", "internal/tui/form.go:Update:107", "entered left/h handler", map[string]interface{}{
				"key":       msg.String(),
				"cursor":    m.cursor,
				"fieldType": fieldType,
			})
			// #endregion
			// For select and confirm fields
			field := &m.fields[m.cursor]
			switch field.prompt.Type {
			case template.PromptSelect:
				if field.selectCursor > 0 {
					field.selectCursor--
				}
			case template.PromptConfirm:
				field.confirmValue = !field.confirmValue
			}
			return m, nil
		case "right", "l":
			// #region agent log
			writeDebugLog("repro-1", "H1", "internal/tui/form.go:Update:119", "entered right/l handler", map[string]interface{}{
				"key":       msg.String(),
				"cursor":    m.cursor,
				"fieldType": fieldType,
			})
			// #endregion
			field := &m.fields[m.cursor]
			switch field.prompt.Type {
			case template.PromptSelect:
				if field.selectCursor < len(field.selectOptions)-1 {
					field.selectCursor++
				}
			case template.PromptConfirm:
				field.confirmValue = !field.confirmValue
			}
			return m, nil
		}
	}

	// Update the focused text input
	if m.cursor < len(m.fields) && m.fields[m.cursor].prompt.Type == template.PromptText {
		var cmd tea.Cmd
		before := m.fields[m.cursor].textInput.Value()
		m.fields[m.cursor].textInput, cmd = m.fields[m.cursor].textInput.Update(msg)
		after := m.fields[m.cursor].textInput.Value()
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			// #region agent log
			writeDebugLog("repro-1", "H2", "internal/tui/form.go:Update:136", "text input update result", map[string]interface{}{
				"key":          keyMsg.String(),
				"beforeLen":    len(before),
				"afterLen":     len(after),
				"valueChanged": before != after,
			})
			// #endregion
		}
		return m, cmd
	}

	return m, nil
}

func (m FormModel) nextField() (FormModel, tea.Cmd) {
	// Blur current text input
	if m.fields[m.cursor].prompt.Type == template.PromptText {
		m.fields[m.cursor].textInput.Blur()
	}

	m.cursor++
	if m.cursor >= len(m.fields) {
		m.cursor = len(m.fields) - 1
	}

	// Focus new text input
	if m.fields[m.cursor].prompt.Type == template.PromptText {
		return m, m.fields[m.cursor].textInput.Focus()
	}
	return m, nil
}

func (m FormModel) prevField() (FormModel, tea.Cmd) {
	if m.fields[m.cursor].prompt.Type == template.PromptText {
		m.fields[m.cursor].textInput.Blur()
	}

	m.cursor--
	if m.cursor < 0 {
		m.cursor = 0
	}

	if m.fields[m.cursor].prompt.Type == template.PromptText {
		return m, m.fields[m.cursor].textInput.Focus()
	}
	return m, nil
}

func (m FormModel) collectAnswers() map[string]interface{} {
	answers := make(map[string]interface{})
	for _, field := range m.fields {
		switch field.prompt.Type {
		case template.PromptText:
			answers[field.prompt.Name] = field.textInput.Value()
		case template.PromptSelect:
			if len(field.selectOptions) > 0 {
				answers[field.prompt.Name] = field.selectOptions[field.selectCursor].Value
			}
		case template.PromptConfirm:
			answers[field.prompt.Name] = field.confirmValue
		}
	}
	return answers
}

func (m FormModel) View() string {
	var b strings.Builder

	// Header
	header := headerStyle.Render(fmt.Sprintf("  %s — Configure", m.manifest.Name))
	b.WriteString(header)
	b.WriteString("\n\n")

	// Fields
	for i, field := range m.fields {
		isFocused := i == m.cursor
		label := field.prompt.Label + ":"
		if isFocused {
			label = focusedStyle.Render(label)
		} else {
			label = labelStyle.Render(label)
		}

		b.WriteString(fmt.Sprintf("  %s  ", label))

		switch field.prompt.Type {
		case template.PromptText:
			b.WriteString(field.textInput.View())

		case template.PromptSelect:
			for j, opt := range field.selectOptions {
				if j == field.selectCursor {
					if isFocused {
						b.WriteString(focusedStyle.Render(fmt.Sprintf("[%s]", opt.Label)))
					} else {
						b.WriteString(fmt.Sprintf("[%s]", opt.Label))
					}
				} else {
					b.WriteString(mutedStyle.Render(fmt.Sprintf(" %s ", opt.Label)))
				}
				b.WriteString(" ")
			}

		case template.PromptConfirm:
			if field.confirmValue {
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
		}

		b.WriteString("\n")
	}

	// Help
	b.WriteString(helpStyle.Render("\n  tab/↓ next • shift+tab/↑ prev • ←/→ change • enter confirm • esc back"))

	return b.String()
}

func writeDebugLog(runID, hypothesisID, location, message string, data map[string]interface{}) {
	entry := map[string]interface{}{
		"sessionId":    "a1bd86",
		"runId":        runID,
		"hypothesisId": hypothesisID,
		"location":     location,
		"message":      message,
		"data":         data,
		"timestamp":    timeNowMillis(),
	}
	payload, err := json.Marshal(entry)
	if err != nil {
		return
	}
	f, err := os.OpenFile("/workspaces/sloth-incubator/.cursor/debug-a1bd86.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(payload, '\n'))
}

func timeNowMillis() int64 {
	return time.Now().UnixMilli()
}
