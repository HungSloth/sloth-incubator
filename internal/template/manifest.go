package template

// PromptType represents the type of a template prompt
type PromptType string

const (
	PromptText    PromptType = "text"
	PromptSelect  PromptType = "select"
	PromptConfirm PromptType = "confirm"
)

// PromptOption represents a selectable option in a prompt
type PromptOption struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
}

// UnmarshalYAML allows PromptOption to be unmarshalled from either a string or a map
func (po *PromptOption) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try string first
	var s string
	if err := unmarshal(&s); err == nil {
		po.Label = s
		po.Value = s
		return nil
	}
	// Try map
	type plain PromptOption
	var p plain
	if err := unmarshal(&p); err != nil {
		return err
	}
	*po = PromptOption(p)
	return nil
}

// Prompt represents a single prompt in a template manifest
type Prompt struct {
	Name     string         `yaml:"name"`
	Label    string         `yaml:"label"`
	Type     PromptType     `yaml:"type"`
	Default  interface{}    `yaml:"default"`
	Options  []PromptOption `yaml:"options"`
	Required bool           `yaml:"required"`
	Validate string         `yaml:"validate"`
}

// FileRule defines a conditional file inclusion rule
type FileRule struct {
	Src    string `yaml:"src"`
	Always bool   `yaml:"always"`
	When   string `yaml:"when"`
}

// DevcontainerFeatures holds conditional devcontainer feature lists
type DevcontainerFeatures struct {
	Always []string `yaml:"always"`
}

// DevcontainerConfig holds devcontainer configuration
type DevcontainerConfig struct {
	BaseImage string               `yaml:"base_image"`
	Features  DevcontainerFeatures `yaml:"features"`
}

// HooksConfig holds hook configuration
type HooksConfig struct {
	PostCreate string `yaml:"post_create"`
}

// TemplateManifest represents a template.yaml file
type TemplateManifest struct {
	Name         string             `yaml:"name"`
	Version      string             `yaml:"version"`
	Description  string             `yaml:"description"`
	Author       string             `yaml:"author"`
	Prompts      []Prompt           `yaml:"prompts"`
	Files        []FileRule         `yaml:"files"`
	Devcontainer DevcontainerConfig `yaml:"devcontainer"`
	Hooks        HooksConfig        `yaml:"hooks"`
}
