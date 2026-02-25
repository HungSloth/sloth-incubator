package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/HungSloth/sloth-incubator/internal/config"
	"github.com/HungSloth/sloth-incubator/internal/container"
	"github.com/HungSloth/sloth-incubator/internal/preview"
	"github.com/HungSloth/sloth-incubator/internal/template"
	"github.com/HungSloth/sloth-incubator/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "incubator",
		Short: "Sloth Incubator — scaffold new projects with ease",
		Long:  "A CLI/TUI tool that standardizes how projects are created. Pick a template, answer a few questions, and get a fully scaffolded project.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return launchTUI()
		},
	}

	newCmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return launchTUI()
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available templates",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.Load()
			manifests := loadAllTemplates(cfg)
			fmt.Printf("%-15s %s\n", "NAME", "DESCRIPTION")
			for _, m := range manifests {
				fmt.Printf("%-15s %s\n", m.Name, m.Description)
			}
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("incubator %s\n", version)
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update templates and check for binary updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Refresh templates
			cacheDir := config.ConfigDir()
			loader := template.NewLoader(cacheDir, cfg.TemplateRepo)
			fmt.Println("Refreshing templates...")
			if err := loader.FetchTemplates(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to refresh templates: %v\n", err)
			} else {
				cache := template.NewCache(cacheDir)
				cache.MarkFetched()
				fmt.Println("Templates updated.")
			}

			// Check for binary update
			fmt.Println("\nCheck GitHub Releases for the latest binary version:")
			fmt.Println("  https://github.com/HungSloth/sloth-incubator/releases")

			return nil
		},
	}

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Edit configuration interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			show, _ := cmd.Flags().GetBool("show")
			if show {
				fmt.Printf("Config file: %s\n\n", config.ConfigPath())
				fmt.Print(cfg.String())
				return nil
			}

			p := tea.NewProgram(tui.NewConfigModel(cfg), tea.WithAltScreen())
			_, err = p.Run()
			return err
		},
	}
	configCmd.Flags().Bool("show", false, "Show current config without editing")

	addRepoCmd := &cobra.Command{
		Use:   "add-repo [url]",
		Short: "Add a community template repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.AddTemplateRepo(args[0])
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Printf("Added template repo: %s\n", args[0])
			return nil
		},
	}

	createTemplateCmd := &cobra.Command{
		Use:   "create-template [name]",
		Short: "Create a local template scaffold",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			localDir := cfg.GetLocalTemplateDir()
			templateDir, err := template.CreateLocalTemplate(localDir, args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Created local template: %s\n", templateDir)
			fmt.Println("Edit template.yaml and files/ to customize it.")
			return nil
		},
	}

	previewCmd := &cobra.Command{
		Use:   "preview [project-dir]",
		Short: "Start a local noVNC preview session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir := "."
			if len(args) == 1 {
				projectDir = args[0]
			}
			absDir, err := filepath.Abs(projectDir)
			if err != nil {
				return fmt.Errorf("resolving project directory: %w", err)
			}

			cfg, err := preview.LoadConfig(absDir)
			if err != nil {
				return err
			}

			url, err := preview.Start(absDir, cfg)
			if err != nil {
				return err
			}

			fmt.Printf("Preview started: %s\n", url)
			if err := preview.OpenBrowser(url); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not open browser automatically: %v\n", err)
				fmt.Printf("Open this URL manually: %s\n", url)
			}

			return nil
		},
	}

	var cleanList bool
	var cleanStopped bool
	var cleanAll bool
	var cleanDryRun bool
	var cleanVolumes bool

	cleanCmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean up devcontainers created for projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			modeCount := 0
			for _, enabled := range []bool{cleanList, cleanStopped, cleanAll} {
				if enabled {
					modeCount++
				}
			}
			if modeCount > 1 {
				return fmt.Errorf("use only one of --list, --stopped, or --all")
			}

			containers, err := container.List()
			if err != nil {
				return err
			}

			if cleanList {
				printDevcontainers(containers)
				return nil
			}

			if len(containers) == 0 {
				fmt.Println("No devcontainers found.")
				return nil
			}

			if cleanStopped {
				targets := filterStopped(containers)
				return cleanupContainers(targets, cleanVolumes, cleanDryRun, false)
			}

			if cleanAll {
				return cleanupContainers(containers, cleanVolumes, cleanDryRun, true)
			}

			selected, cancelled, err := tui.RunCleanSelection(containers)
			if err != nil {
				return err
			}
			if cancelled {
				fmt.Println("Cleanup cancelled.")
				return nil
			}
			return cleanupContainers(selected, cleanVolumes, cleanDryRun, true)
		},
	}
	cleanCmd.Flags().BoolVar(&cleanList, "list", false, "List devcontainers and exit")
	cleanCmd.Flags().BoolVar(&cleanStopped, "stopped", false, "Remove only stopped devcontainers")
	cleanCmd.Flags().BoolVar(&cleanAll, "all", false, "Stop and remove all devcontainers")
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Show planned actions without making changes")
	cleanCmd.Flags().BoolVar(&cleanVolumes, "volumes", false, "Also remove container volumes")

	rootCmd.AddCommand(newCmd, listCmd, versionCmd, updateCmd, configCmd, addRepoCmd, createTemplateCmd, previewCmd, cleanCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func launchTUI() error {
	cfg, _ := config.Load()
	manifests := loadAllTemplates(cfg)
	p := tea.NewProgram(tui.NewApp(manifests, cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func loadAllTemplates(cfg *config.Config) []*template.TemplateManifest {
	// Start with the built-in template
	manifests := []*template.TemplateManifest{
		template.GetBuiltinManifest(),
	}

	// Try loading from cache
	cacheDir := config.ConfigDir()
	loader := template.NewLoader(cacheDir, cfg.TemplateRepo)
	cache := template.NewCache(cacheDir)

	// Background refresh if stale
	if !cache.NeedsInitialFetch() {
		if remote, err := loader.LoadAllManifests(); err == nil && len(remote) > 0 {
			manifests = append(manifests, remote...)
		}
	}

	if cfg != nil {
		if local, err := template.LoadLocalManifests(cfg.GetLocalTemplateDir()); err == nil && len(local) > 0 {
			manifests = append(manifests, local...)
		}
	}

	return manifests
}

func printDevcontainers(containers []container.DevContainer) {
	if len(containers) == 0 {
		fmt.Println("No devcontainers found.")
		return
	}

	fmt.Printf("%-14s %-22s %-24s %s\n", "CONTAINER ID", "NAME", "STATUS", "PROJECT")
	for _, c := range containers {
		id := c.ID
		if len(id) > 12 {
			id = id[:12]
		}
		fmt.Printf("%-14s %-22s %-24s %s\n", id, c.Name, c.Status, c.ProjectDir)
	}
}

func filterStopped(containers []container.DevContainer) []container.DevContainer {
	stopped := make([]container.DevContainer, 0, len(containers))
	for _, c := range containers {
		if !isRunningStatus(c.Status) {
			stopped = append(stopped, c)
		}
	}
	return stopped
}

func cleanupContainers(containers []container.DevContainer, removeVolumes, dryRun, stopRunning bool) error {
	if len(containers) == 0 {
		fmt.Println("No matching devcontainers found.")
		return nil
	}

	removed := 0
	stopped := 0
	for _, c := range containers {
		running := isRunningStatus(c.Status)
		if running && stopRunning {
			if dryRun {
				fmt.Printf("[dry-run] docker stop %s (%s)\n", c.ID, c.Name)
			} else {
				if err := container.Stop(c.ID); err != nil {
					return err
				}
			}
			stopped++
		}

		if dryRun {
			if removeVolumes {
				fmt.Printf("[dry-run] docker rm -v %s (%s)\n", c.ID, c.Name)
			} else {
				fmt.Printf("[dry-run] docker rm %s (%s)\n", c.ID, c.Name)
			}
		} else {
			if err := container.Remove(c.ID, removeVolumes); err != nil {
				return err
			}
		}
		removed++
	}

	if dryRun {
		fmt.Printf("Dry run complete: %d container(s) would be removed (%d would be stopped first).\n", removed, stopped)
		return nil
	}
	fmt.Printf("Cleanup complete: removed %d container(s), stopped %d running container(s).\n", removed, stopped)
	return nil
}

func isRunningStatus(status string) bool {
	return strings.HasPrefix(status, "Up ") || strings.HasPrefix(status, "Restarting")
}
