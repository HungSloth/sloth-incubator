package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/HungSloth/sloth-incubator/internal/config"
	"github.com/HungSloth/sloth-incubator/internal/template"
	tea "github.com/charmbracelet/bubbletea"
)

// ConfirmModel handles the confirmation screen
type ConfirmModel struct {
	manifest      *template.TemplateManifest
	answers       map[string]interface{}
	files         []string
	initMode      bool
	targetDir     string
	newFiles      []string
	existingFiles []string
}

// NewConfirmModel creates a new confirmation model
func NewConfirmModel(manifest *template.TemplateManifest, answers map[string]interface{}, cfg *config.Config, initMode bool, targetDir string) ConfirmModel {
	// Get the list of files from the selected template source.
	renderer := template.NewRenderer(manifest, answers)
	templateRepo := config.DefaultConfig().TemplateRepo
	if cfg != nil && cfg.TemplateRepo != "" {
		templateRepo = cfg.TemplateRepo
	}
	templateFS, err := template.ResolveTemplateFS(manifest, config.ConfigDir(), templateRepo)
	var files []string
	if err == nil {
		files, _ = renderer.ListFiles(templateFS)
	}

	newFiles := make([]string, 0, len(files))
	existingFiles := make([]string, 0, len(files))
	if initMode && targetDir != "" {
		for _, f := range files {
			targetPath := filepath.Join(targetDir, f)
			if _, err := os.Stat(targetPath); err == nil {
				existingFiles = append(existingFiles, f)
			} else {
				newFiles = append(newFiles, f)
			}
		}
	}

	return ConfirmModel{
		manifest:      manifest,
		answers:       answers,
		files:         files,
		initMode:      initMode,
		targetDir:     targetDir,
		newFiles:      newFiles,
		existingFiles: existingFiles,
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
	headerText := "  Ready to create"
	if m.initMode {
		headerText = "  Ready to initialize"
	}
	header := headerStyle.Render(headerText)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Summary
	projectName := m.getAnswer("project_name", "unnamed")
	b.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Render("Name:"), valueStyle.Render(projectName)))
	b.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Render("Template:"), valueStyle.Render(m.manifest.Name)))

	if vis := m.getAnswer("visibility", ""); vis != "" {
		if createRepo := m.getAnswer("create_github_repo", "true"); createRepo == "true" {
			b.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Render("Visibility:"), valueStyle.Render(vis)))
		}
	}
	if lic := m.getAnswer("license", ""); lic != "" {
		b.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Render("License:"), valueStyle.Render(lic)))
	}

	// Files
	if m.initMode {
		if len(m.newFiles) > 0 {
			b.WriteString(fmt.Sprintf("\n  %s\n", titleStyle.Render("Files to create:")))
			for _, f := range m.newFiles {
				b.WriteString(fmt.Sprintf("    %s\n", mutedStyle.Render(f)))
			}
		}
		if len(m.existingFiles) > 0 {
			b.WriteString(fmt.Sprintf("\n  %s\n", titleStyle.Render("Existing files (skipped):")))
			for _, f := range m.existingFiles {
				b.WriteString(fmt.Sprintf("    %s %s\n", mutedStyle.Render(f), mutedStyle.Render("(skip)")))
			}
		}
	} else if len(m.files) > 0 {
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
