package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	repoOwner = "HungSloth"
	repoName  = "sloth-incubator"
)

// Release represents a GitHub release
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a GitHub release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	HasUpdate      bool
}

// CheckForUpdate checks if a newer version is available
func CheckForUpdate(currentVersion string) (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("checking for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parsing release info: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentClean := strings.TrimPrefix(currentVersion, "v")

	return &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		HasUpdate:      latestVersion != currentClean && currentClean != "dev",
	}, nil
}

// SelfUpdate downloads and installs the latest version
func SelfUpdate(currentVersion string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("fetching release info: %w", err)
	}
	defer resp.Body.Close()

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("parsing release info: %w", err)
	}

	// Find the right asset for this OS/arch
	assetName := fmt.Sprintf("incubator_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no release asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download the asset
	dlResp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("downloading update: %w", err)
	}
	defer dlResp.Body.Close()

	// Write to a temp file
	tmpDir, err := os.MkdirTemp("", "incubator-update-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, assetName)
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	if _, err := io.Copy(f, dlResp.Body); err != nil {
		f.Close()
		return fmt.Errorf("downloading update: %w", err)
	}
	f.Close()

	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting executable path: %w", err)
	}

	fmt.Printf("Downloaded %s\n", release.TagName)
	fmt.Printf("To complete the update, extract %s and replace %s\n", tmpFile, execPath)

	return nil
}

// FormatUpdateBanner returns a formatted update banner string
func FormatUpdateBanner(info *UpdateInfo) string {
	if !info.HasUpdate {
		return ""
	}
	return fmt.Sprintf("Update available: %s -> %s (run incubator update)", info.CurrentVersion, info.LatestVersion)
}
