package template

import (
	"embed"
	"io/fs"
)

//go:embed all:embedded/empty
var embeddedEmpty embed.FS

// GetEmbeddedEmptyTemplate returns the filesystem for the embedded "empty" template
func GetEmbeddedEmptyTemplate() (fs.FS, error) {
	return fs.Sub(embeddedEmpty, "embedded/empty")
}

// GetBuiltinManifest returns the manifest for the built-in "empty" template
func GetBuiltinManifest() *TemplateManifest {
	return &TemplateManifest{
		Name:        "empty",
		Version:     "1.0.0",
		Description: "Blank project with devcontainer",
		Author:      "HungSloth",
		Prompts: []Prompt{
			{
				Name:     "project_name",
				Label:    "Project name",
				Type:     PromptText,
				Required: true,
			},
			{
				Name:    "description",
				Label:   "Description",
				Type:    PromptText,
				Default: "A new project",
			},
			{
				Name:  "visibility",
				Label: "Repo visibility",
				Type:  PromptSelect,
				Options: []PromptOption{
					{Label: "private", Value: "private"},
					{Label: "public", Value: "public"},
				},
				Default: "private",
			},
			{
				Name:  "license",
				Label: "License",
				Type:  PromptSelect,
				Options: []PromptOption{
					{Label: "MIT", Value: "MIT"},
					{Label: "Apache-2.0", Value: "Apache-2.0"},
					{Label: "GPL-3.0", Value: "GPL-3.0"},
					{Label: "none", Value: "none"},
				},
				Default: "MIT",
			},
		},
	}
}
