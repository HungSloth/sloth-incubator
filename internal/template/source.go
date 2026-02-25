package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ResolveTemplateFS resolves the filesystem for either built-in or remote templates.
func ResolveTemplateFS(manifest *TemplateManifest, cacheDir, templateRepo string) (fs.FS, error) {
	if manifest == nil || manifest.IsBuiltin || manifest.SourcePath == "" {
		return GetEmbeddedEmptyTemplate()
	}
	if filepath.IsAbs(manifest.SourcePath) {
		filesDir := filepath.Join(manifest.SourcePath, "files")
		stat, err := os.Stat(filesDir)
		if err != nil || !stat.IsDir() {
			return nil, fmt.Errorf("local template files directory not found: %s", filesDir)
		}
		return os.DirFS(filesDir), nil
	}

	loader := NewLoader(cacheDir, templateRepo)
	return loader.GetTemplateFS(manifest.SourcePath)
}
