package config

import (
	"path/filepath"
	"testing"
)

func TestGetLocalTemplateDirExpandsHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := &Config{LocalTemplateDir: "~/.incubator/local-templates"}
	got := cfg.GetLocalTemplateDir()
	want := filepath.Join(home, ".incubator", "local-templates")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestGetLocalTemplateDirUsesDefaultWhenEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := &Config{}
	got := cfg.GetLocalTemplateDir()
	want := filepath.Join(home, ".incubator", "local-templates")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
