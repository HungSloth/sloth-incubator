package tui

import (
	"testing"

	"github.com/HungSloth/sloth-incubator/internal/template"
	tea "github.com/charmbracelet/bubbletea"
)

func TestFormTextFieldAcceptsRegularRunesExceptMappedKeys(t *testing.T) {
	manifest := &template.TemplateManifest{
		Name: "test",
		Prompts: []template.Prompt{
			{
				Name:  "project_name",
				Label: "Project name",
				Type:  template.PromptText,
			},
		},
	}
	model := NewFormModel(manifest)

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if got := model.fields[0].textInput.Value(); got != "a" {
		t.Fatalf("expected text input to contain 'a', got %q", got)
	}

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if got := model.fields[0].textInput.Value(); got != "ah" {
		t.Fatalf("expected text input to contain 'ah', got %q", got)
	}

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if got := model.fields[0].textInput.Value(); got != "ahl" {
		t.Fatalf("expected text input to contain 'ahl', got %q", got)
	}
}

func TestFormSelectFieldUsesHLForNavigation(t *testing.T) {
	manifest := &template.TemplateManifest{
		Name: "test",
		Prompts: []template.Prompt{
			{
				Name:  "language",
				Label: "Language",
				Type:  template.PromptSelect,
				Options: []template.PromptOption{
					{Label: "Go", Value: "go"},
					{Label: "Rust", Value: "rust"},
				},
			},
		},
	}
	model := NewFormModel(manifest)

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if got := model.fields[0].selectCursor; got != 1 {
		t.Fatalf("expected select cursor to move right to 1, got %d", got)
	}

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if got := model.fields[0].selectCursor; got != 0 {
		t.Fatalf("expected select cursor to move left to 0, got %d", got)
	}
}
