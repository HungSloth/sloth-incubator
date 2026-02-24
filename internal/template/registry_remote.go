package template

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// CommunityTemplate represents a community template repo entry
type CommunityTemplate struct {
	Name        string `json:"name" yaml:"name"`
	Repo        string `json:"repo" yaml:"repo"`
	Description string `json:"description" yaml:"description"`
	Author      string `json:"author" yaml:"author"`
	Verified    bool   `json:"verified" yaml:"verified"`
}

// CommunityRegistry represents the community template registry
type CommunityRegistry struct {
	Templates []CommunityTemplate `json:"templates" yaml:"templates"`
}

// FetchCommunityRegistry fetches the community template registry from GitHub
func FetchCommunityRegistry() (*CommunityRegistry, error) {
	url := "https://raw.githubusercontent.com/HungSloth/sloth-incubator/main/community-templates.json"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching community registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("community registry returned status %d", resp.StatusCode)
	}

	var registry CommunityRegistry
	if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
		return nil, fmt.Errorf("parsing community registry: %w", err)
	}

	return &registry, nil
}
