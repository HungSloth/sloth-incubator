package tui

import (
	"github.com/HungSloth/sloth-incubator/internal/config"
	"github.com/HungSloth/sloth-incubator/internal/template"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents the current TUI screen
type Screen int

const (
	ScreenPicker Screen = iota
	ScreenForm
	ScreenConfirm
	ScreenProgress
	ScreenDone
)

// App is the root Bubbletea model
type App struct {
	screen   Screen
	picker   PickerModel
	form     FormModel
	confirm  ConfirmModel
	progress ProgressModel
	done     DoneModel
	width    int
	height   int

	// Shared state
	cfg              *config.Config
	selectedTemplate *template.TemplateManifest
	answers          map[string]interface{}
	projectDir       string
	repoURL          string
	quitting         bool
}

// NewApp creates a new App model
func NewApp(manifests []*template.TemplateManifest, cfg *config.Config) App {
	return App{
		screen:  ScreenPicker,
		picker:  NewPickerModel(manifests),
		cfg:     cfg,
		answers: make(map[string]interface{}),
	}
}

func (a App) Init() tea.Cmd {
	return a.picker.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			a.quitting = true
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	}

	switch msg := msg.(type) {
	case templateSelectedMsg:
		a.selectedTemplate = msg.manifest
		a.form = NewFormModel(msg.manifest)
		a.screen = ScreenForm
		return a, a.form.Init()

	case formCompletedMsg:
		a.answers = msg.answers
		a.confirm = NewConfirmModel(a.selectedTemplate, a.answers)
		a.screen = ScreenConfirm
		return a, nil

	case formBackMsg:
		a.screen = ScreenPicker
		return a, nil

	case confirmProceedMsg:
		a.progress = NewProgressModel(a.selectedTemplate, a.answers, a.cfg)
		a.screen = ScreenProgress
		return a, a.progress.Init()

	case confirmBackMsg:
		a.screen = ScreenForm
		return a, nil

	case progressDoneMsg:
		a.projectDir = msg.projectDir
		a.repoURL = msg.repoURL
		a.done = NewDoneModel(a.projectDir, a.repoURL)
		a.screen = ScreenDone
		return a, nil

	case quitMsg:
		a.quitting = true
		return a, tea.Quit
	}

	// Delegate to the active screen
	var cmd tea.Cmd
	switch a.screen {
	case ScreenPicker:
		a.picker, cmd = a.picker.Update(msg)
	case ScreenForm:
		a.form, cmd = a.form.Update(msg)
	case ScreenConfirm:
		a.confirm, cmd = a.confirm.Update(msg)
	case ScreenProgress:
		a.progress, cmd = a.progress.Update(msg)
	case ScreenDone:
		a.done, cmd = a.done.Update(msg)
	}

	return a, cmd
}

func (a App) View() string {
	if a.quitting {
		return ""
	}

	var content string
	switch a.screen {
	case ScreenPicker:
		content = a.picker.View()
	case ScreenForm:
		content = a.form.View()
	case ScreenConfirm:
		content = a.confirm.View()
	case ScreenProgress:
		content = a.progress.View()
	case ScreenDone:
		content = a.done.View()
	}

	return appStyle.Render(content)
}

// Message types for screen transitions
type templateSelectedMsg struct {
	manifest *template.TemplateManifest
}

type formCompletedMsg struct {
	answers map[string]interface{}
}

type formBackMsg struct{}

type confirmProceedMsg struct{}

type confirmBackMsg struct{}

type progressDoneMsg struct {
	projectDir string
	repoURL    string
}

type quitMsg struct{}
