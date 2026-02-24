package tui

import (
	"fmt"
	"strings"

	"github.com/HungSloth/sloth-incubator/internal/template"
	tea "github.com/charmbracelet/bubbletea"
)

// ConfirmModel handles the confirmation screen
type ConfirmModel struct {
	manifest *template.TemplateManifest
	answers  map[string]interface{}
	files    []string
}

// NewConfirmModel creates a new confirmation model
func NewConfirmModel(manifest *template.TemplateManifest, answers map[string]interface{}) ConfirmModel {
	// Get the list of files from the embedded template
	renderer := template.NewRenderer(manifest, answers)
	templateFS, err := template.GetEmbeddedEmptyTemplate()
	var files []string
	if err == nil {
		files, _ = renderer.ListFiles(templateFS)
	}

	return ConfirmModel{
		manifest: manifest,
		answers:  answers,
		files:    files,
	}
}

func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, func() tea.Msg { return confirmProceedMsg{} }
		case "esc":
			return m, func() tea.Msg { return confirmBackMsg{} }
		case "q":
			return m, func() tea.Msg { return quitMsg{} }
		}
	}
	return m, nil
}

func (m ConfirmModel) View() string {
	var b strings.Builder

	// Header
	header := headerStyle.Render("  Ready to create")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Summary
	projectName := m.getAnswer("project_name", "unnamed")
	b.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Render("Name:"), valueStyle.Render(projectName)))
	b.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Render("Template:"), valueStyle.Render(m.manifest.Name)))

	if vis := m.getAnswer("visibility", ""); vis != "" {
		b.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Render("Visibility:"), valueStyle.Render(vis)))
	}
	if lic := m.getAnswer("license", ""); lic != "" {
		b.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Render("License:"), valueStyle.Render(lic)))
	}

	// Files
	if len(m.files) > 0 {
		b.WriteString(fmt.Sprintf("\n  %s\n", titleStyle.Render("Files to create:")))
		for _, f := range m.files {
			b.WriteString(fmt.Sprintf("    %s\n", mutedStyle.Render(f)))
		}
	}

	// Help
	b.WriteString(helpStyle.Render("\n  enter create • esc back • q cancel"))

	return b.String()
}

func (m ConfirmModel) getAnswer(key, fallback string) string {
	if val, ok := m.answers[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return fallback
}
