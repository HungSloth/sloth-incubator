package tui

import (
	"fmt"
	"strings"

	"github.com/HungSloth/sloth-incubator/internal/config"
	"github.com/HungSloth/sloth-incubator/internal/template"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type templateCreatorStep int

const (
	templateCreatorStepName templateCreatorStep = iota
	templateCreatorStepDescription
	templateCreatorStepBase
	templateCreatorStepSoftware
	templateCreatorStepTools
	templateCreatorStepReview
	templateCreatorStepDone
)

type optionItem struct {
	label string
	value string
	desc  string
}

// TemplateCreatorModel handles local template generation via wizard steps.
type TemplateCreatorModel struct {
	cfg *config.Config

	step templateCreatorStep

	nameInput        textinput.Model
	descriptionInput textinput.Model

	baseOptions     []optionItem
	softwareOptions []optionItem
	toolOptions     []optionItem

	baseCursor     int
	softwareCursor int
	toolsCursor    int
	selectedTools  map[string]bool

	errorMsg    string
	createdPath string
}

func NewTemplateCreatorModel(cfg *config.Config) TemplateCreatorModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "my-template"
	nameInput.Focus()

	descriptionInput := textinput.New()
	descriptionInput.Placeholder = "Local template scaffold"

	return TemplateCreatorModel{
		cfg:              cfg,
		step:             templateCreatorStepName,
		nameInput:        nameInput,
		descriptionInput: descriptionInput,
		baseOptions: []optionItem{
			{label: "Empty", value: "empty", desc: "Minimal starter template"},
			{label: "Basic CLI", value: "basic-cli", desc: "CLI-friendly starter layout"},
			{label: "Basic Web", value: "basic-web", desc: "Web app starter layout"},
		},
		softwareOptions: []optionItem{
			{label: "Go", value: "go", desc: "Go-first defaults"},
			{label: "Node", value: "node", desc: "Node.js-friendly defaults"},
			{label: "Python", value: "python", desc: "Python-friendly defaults"},
		},
		toolOptions: []optionItem{
			{label: "Git", value: "git", desc: "Include .gitignore starter"},
			{label: "Devcontainer", value: "devcontainer", desc: "Include .devcontainer setup"},
			{label: "Preview", value: "preview", desc: "Include .incubator/preview assets"},
		},
		selectedTools: map[string]bool{
			"git":          true,
			"devcontainer": true,
			"preview":      true,
		},
	}
}

func (m TemplateCreatorModel) Init() tea.Cmd {
	return m.nameInput.Focus()
}

func (m TemplateCreatorModel) Update(msg tea.Msg) (TemplateCreatorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.errorMsg = ""
		switch msg.String() {
		case "q":
			return m, func() tea.Msg { return quitMsg{} }
		case "esc":
			return m.handleBack()
		}

		switch m.step {
		case templateCreatorStepName:
			return m.updateNameStep(msg)
		case templateCreatorStepDescription:
			return m.updateDescriptionStep(msg)
		case templateCreatorStepBase:
			return m.updateBaseStep(msg)
		case templateCreatorStepSoftware:
			return m.updateSoftwareStep(msg)
		case templateCreatorStepTools:
			return m.updateToolsStep(msg)
		case templateCreatorStepReview:
			return m.updateReviewStep(msg)
		case templateCreatorStepDone:
			if msg.String() == "enter" {
				return m, func() tea.Msg { return templateCreatorDoneMsg{} }
			}
		}
	}
	return m, nil
}

func (m TemplateCreatorModel) handleBack() (TemplateCreatorModel, tea.Cmd) {
	switch m.step {
	case templateCreatorStepName:
		return m, func() tea.Msg { return templateCreatorBackMsg{} }
	case templateCreatorStepDescription:
		m.descriptionInput.Blur()
		m.nameInput.Focus()
		m.step = templateCreatorStepName
	case templateCreatorStepBase:
		m.descriptionInput.Focus()
		m.step = templateCreatorStepDescription
	case templateCreatorStepSoftware:
		m.step = templateCreatorStepBase
	case templateCreatorStepTools:
		m.step = templateCreatorStepSoftware
	case templateCreatorStepReview:
		m.step = templateCreatorStepTools
	case templateCreatorStepDone:
		return m, func() tea.Msg { return templateCreatorDoneMsg{} }
	}
	return m, nil
}

func (m TemplateCreatorModel) updateNameStep(msg tea.KeyMsg) (TemplateCreatorModel, tea.Cmd) {
	if msg.String() == "enter" {
		if strings.TrimSpace(m.nameInput.Value()) == "" {
			m.errorMsg = "Template name is required."
			return m, nil
		}
		m.nameInput.Blur()
		m.descriptionInput.Focus()
		m.step = templateCreatorStepDescription
		return m, nil
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m TemplateCreatorModel) updateDescriptionStep(msg tea.KeyMsg) (TemplateCreatorModel, tea.Cmd) {
	if msg.String() == "enter" {
		m.descriptionInput.Blur()
		m.step = templateCreatorStepBase
		return m, nil
	}

	var cmd tea.Cmd
	m.descriptionInput, cmd = m.descriptionInput.Update(msg)
	return m, cmd
}

func (m TemplateCreatorModel) updateBaseStep(msg tea.KeyMsg) (TemplateCreatorModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.baseCursor > 0 {
			m.baseCursor--
		}
	case "down", "j":
		if m.baseCursor < len(m.baseOptions)-1 {
			m.baseCursor++
		}
	case "enter":
		m.step = templateCreatorStepSoftware
	}
	return m, nil
}

func (m TemplateCreatorModel) updateSoftwareStep(msg tea.KeyMsg) (TemplateCreatorModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.softwareCursor > 0 {
			m.softwareCursor--
		}
	case "down", "j":
		if m.softwareCursor < len(m.softwareOptions)-1 {
			m.softwareCursor++
		}
	case "enter":
		m.step = templateCreatorStepTools
	}
	return m, nil
}

func (m TemplateCreatorModel) updateToolsStep(msg tea.KeyMsg) (TemplateCreatorModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.toolsCursor > 0 {
			m.toolsCursor--
		}
	case "down", "j":
		if m.toolsCursor < len(m.toolOptions)-1 {
			m.toolsCursor++
		}
	case " ":
		tool := m.toolOptions[m.toolsCursor].value
		m.selectedTools[tool] = !m.selectedTools[tool]
	case "enter":
		m.step = templateCreatorStepReview
	}
	return m, nil
}

func (m TemplateCreatorModel) updateReviewStep(msg tea.KeyMsg) (TemplateCreatorModel, tea.Cmd) {
	if msg.String() != "enter" {
		return m, nil
	}

	localDir := config.DefaultConfig().GetLocalTemplateDir()
	if m.cfg != nil {
		localDir = m.cfg.GetLocalTemplateDir()
	}

	resultPath, err := template.CreateLocalTemplateFromWizard(localDir, template.LocalTemplateWizardOptions{
		Name:        strings.TrimSpace(m.nameInput.Value()),
		Description: strings.TrimSpace(m.descriptionInput.Value()),
		Base:        m.baseOptions[m.baseCursor].value,
		Software:    m.softwareOptions[m.softwareCursor].value,
		Tools:       m.selectedToolValues(),
	})
	if err != nil {
		m.errorMsg = err.Error()
		return m, nil
	}

	m.createdPath = resultPath
	m.step = templateCreatorStepDone
	return m, nil
}

func (m TemplateCreatorModel) selectedToolValues() []string {
	selected := make([]string, 0, len(m.toolOptions))
	for _, item := range m.toolOptions {
		if m.selectedTools[item.value] {
			selected = append(selected, item.value)
		}
	}
	return selected
}

func (m TemplateCreatorModel) View() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("  Template Creator Wizard"))
	b.WriteString("\n\n")

	switch m.step {
	case templateCreatorStepName:
		b.WriteString("  Template name:\n")
		b.WriteString(fmt.Sprintf("  %s\n", m.nameInput.View()))
		b.WriteString(helpStyle.Render("\n  enter continue • esc back • q quit"))
	case templateCreatorStepDescription:
		b.WriteString("  Description (optional):\n")
		b.WriteString(fmt.Sprintf("  %s\n", m.descriptionInput.View()))
		b.WriteString(helpStyle.Render("\n  enter continue • esc back • q quit"))
	case templateCreatorStepBase:
		b.WriteString("  Choose a base:\n\n")
		b.WriteString(m.renderSingleSelectList(m.baseOptions, m.baseCursor))
		b.WriteString(helpStyle.Render("\n  ↑/↓ navigate • enter continue • esc back • q quit"))
	case templateCreatorStepSoftware:
		b.WriteString("  Choose primary software:\n\n")
		b.WriteString(m.renderSingleSelectList(m.softwareOptions, m.softwareCursor))
		b.WriteString(helpStyle.Render("\n  ↑/↓ navigate • enter continue • esc back • q quit"))
	case templateCreatorStepTools:
		b.WriteString("  Select tools:\n\n")
		b.WriteString(m.renderToolList())
		b.WriteString(helpStyle.Render("\n  ↑/↓ navigate • space toggle • enter review • esc back • q quit"))
	case templateCreatorStepReview:
		b.WriteString(m.renderReview())
		b.WriteString(helpStyle.Render("\n  enter create template • esc back • q quit"))
	case templateCreatorStepDone:
		b.WriteString(successStyle.Render("  Template created successfully.\n\n"))
		b.WriteString(fmt.Sprintf("  %s\n", valueStyle.Render(m.createdPath)))
		b.WriteString(helpStyle.Render("\n  enter return to main menu • esc return"))
	}

	if m.errorMsg != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("  Error: " + m.errorMsg))
	}

	return b.String()
}

func (m TemplateCreatorModel) renderSingleSelectList(options []optionItem, cursor int) string {
	var b strings.Builder
	for i, opt := range options {
		cursorPrefix := "  "
		style := inactiveItemStyle
		if i == cursor {
			cursorPrefix = "> "
			style = activeItemStyle
		}
		b.WriteString(fmt.Sprintf("%s%s  %s\n", cursorPrefix, style.Render(opt.label), mutedStyle.Render(opt.desc)))
	}
	return b.String()
}

func (m TemplateCreatorModel) renderToolList() string {
	var b strings.Builder
	for i, opt := range m.toolOptions {
		cursorPrefix := "  "
		if i == m.toolsCursor {
			cursorPrefix = "> "
		}

		check := "[ ]"
		if m.selectedTools[opt.value] {
			check = "[x]"
		}

		line := fmt.Sprintf("%s%s %s  %s", cursorPrefix, check, opt.label, mutedStyle.Render(opt.desc))
		if i == m.toolsCursor {
			b.WriteString(activeItemStyle.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m TemplateCreatorModel) renderReview() string {
	base := m.baseOptions[m.baseCursor].label
	software := m.softwareOptions[m.softwareCursor].label
	tools := m.selectedToolValues()
	if len(tools) == 0 {
		tools = []string{"none"}
	}

	var b strings.Builder
	b.WriteString("  Review template settings:\n\n")
	b.WriteString(fmt.Sprintf("  Name:        %s\n", valueStyle.Render(strings.TrimSpace(m.nameInput.Value()))))
	b.WriteString(fmt.Sprintf("  Description: %s\n", valueStyle.Render(strings.TrimSpace(m.descriptionInput.Value()))))
	b.WriteString(fmt.Sprintf("  Base:        %s\n", valueStyle.Render(base)))
	b.WriteString(fmt.Sprintf("  Software:    %s\n", valueStyle.Render(software)))
	b.WriteString(fmt.Sprintf("  Tools:       %s\n", valueStyle.Render(strings.Join(tools, ", "))))
	return b.String()
}

type templateCreatorBackMsg struct{}

type templateCreatorDoneMsg struct{}
