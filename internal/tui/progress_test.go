package tui

import (
	"testing"

	"github.com/HungSloth/sloth-incubator/internal/template"
)

func TestNewProgressModelSkipsGitHubStepsWhenDisabled(t *testing.T) {
	manifest := template.GetBuiltinManifest()
	answers := map[string]interface{}{
		"project_name":        "demo",
		"create_github_repo": false,
	}

	model := NewProgressModel(manifest, answers, nil, false, "")
	if len(model.steps) != 3 {
		t.Fatalf("expected 3 steps when GitHub is disabled, got %d", len(model.steps))
	}

	for _, step := range model.steps {
		if step.Name == stepCreateGitHubRepo || step.Name == stepPushToOrigin {
			t.Fatalf("unexpected GitHub step found when disabled: %s", step.Name)
		}
	}
}

func TestNewProgressModelIncludesGitHubStepsByDefault(t *testing.T) {
	manifest := template.GetBuiltinManifest()
	answers := map[string]interface{}{
		"project_name": "demo",
	}

	model := NewProgressModel(manifest, answers, nil, false, "")
	if len(model.steps) != 5 {
		t.Fatalf("expected 5 steps by default, got %d", len(model.steps))
	}

	foundCreate := false
	foundPush := false
	for _, step := range model.steps {
		if step.Name == stepCreateGitHubRepo {
			foundCreate = true
		}
		if step.Name == stepPushToOrigin {
			foundPush = true
		}
	}
	if !foundCreate || !foundPush {
		t.Fatalf("expected GitHub steps to be present by default")
	}
}
