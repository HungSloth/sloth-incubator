package tui

import (
	"github.com/HungSloth/sloth-incubator/internal/config"
	"github.com/HungSloth/sloth-incubator/internal/template"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents the current TUI screen
type Screen int

const (
	ScreenMainMenu Screen = iota
	ScreenPicker
	ScreenForm
	ScreenConfirm
	ScreenProgress
	ScreenDone
	ScreenTemplateCreator
)

// App is the root Bubbletea model
type App struct {
	screen          Screen
	menu            MenuModel
	picker          PickerModel
	form            FormModel
	confirm         ConfirmModel
	progress        ProgressModel
	done            DoneModel
	templateCreator TemplateCreatorModel
	width           int
	height          int

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
		screen:          ScreenMainMenu,
		menu:            NewMenuModel(),
		picker:          NewPickerModel(manifests),
		templateCreator: NewTemplateCreatorModel(cfg),
		cfg:             cfg,
		answers:         make(map[string]interface{}),
	}
}

func (a App) Init() tea.Cmd {
	return a.menu.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if a.screen == ScreenForm {
			// #region agent log
			writeDebugLog("repro-1", "H4", "internal/tui/app.go:Update:62", "app received key while on form screen", map[string]interface{}{
				"key":    msg.String(),
				"screen": "form",
			})
			// #endregion
		}
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
	case menuCreateProjectMsg:
		a.screen = ScreenPicker
		return a, a.picker.Init()

	case menuCreateTemplateMsg:
		a.templateCreator = NewTemplateCreatorModel(a.cfg)
		a.screen = ScreenTemplateCreator
		return a, a.templateCreator.Init()

	case templateSelectedMsg:
		a.selectedTemplate = msg.manifest
		a.form = NewFormModel(msg.manifest)
		a.screen = ScreenForm
		return a, a.form.Init()

	case formCompletedMsg:
		a.answers = msg.answers
		a.confirm = NewConfirmModel(a.selectedTemplate, a.answers, a.cfg)
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

	case templateCreatorBackMsg:
		a.screen = ScreenMainMenu
		return a, nil

	case templateCreatorDoneMsg:
		a.screen = ScreenMainMenu
		return a, nil
	}

	// Delegate to the active screen
	var cmd tea.Cmd
	switch a.screen {
	case ScreenMainMenu:
		a.menu, cmd = a.menu.Update(msg)
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
	case ScreenTemplateCreator:
		a.templateCreator, cmd = a.templateCreator.Update(msg)
	}

	return a, cmd
}

func (a App) View() string {
	if a.quitting {
		return ""
	}

	var content string
	switch a.screen {
	case ScreenMainMenu:
		content = a.menu.View()
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
	case ScreenTemplateCreator:
		content = a.templateCreator.View()
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

type menuCreateProjectMsg struct{}

type menuCreateTemplateMsg struct{}
