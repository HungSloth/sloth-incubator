package template

import "io/fs"

// ResolveTemplateFS resolves the filesystem for either built-in or remote templates.
func ResolveTemplateFS(manifest *TemplateManifest, cacheDir, templateRepo string) (fs.FS, error) {
	if manifest == nil || manifest.IsBuiltin || manifest.SourcePath == "" {
		return GetEmbeddedEmptyTemplate()
	}

	loader := NewLoader(cacheDir, templateRepo)
	return loader.GetTemplateFS(manifest.SourcePath)
}
