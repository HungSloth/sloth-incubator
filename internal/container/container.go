package container

import (
	"fmt"
	"os/exec"
	"strings"
)

const devcontainerFolderLabel = "devcontainer.local_folder"

var (
	runCommand = runCmd
	runOutput  = runCmdOutput
)

// DevContainer describes a devcontainer discovered via Docker labels.
type DevContainer struct {
	ID         string
	Name       string
	Status     string
	CreatedAt  string
	ProjectDir string
}

// List returns all containers that have the devcontainer.local_folder label.
func List() ([]DevContainer, error) {
	out, err := runOutput("", "docker", "ps", "-a", "--filter", "label="+devcontainerFolderLabel, "--format", "{{.ID}}\t{{.Names}}\t{{.Status}}\t{{.CreatedAt}}\t{{.Label \"devcontainer.local_folder\"}}")
	if err != nil {
		return nil, fmt.Errorf("listing devcontainers: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	containers := make([]DevContainer, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 5)
		if len(parts) < 5 {
			continue
		}

		containers = append(containers, DevContainer{
			ID:         strings.TrimSpace(parts[0]),
			Name:       strings.TrimSpace(parts[1]),
			Status:     strings.TrimSpace(parts[2]),
			CreatedAt:  strings.TrimSpace(parts[3]),
			ProjectDir: strings.TrimSpace(parts[4]),
		})
	}

	return containers, nil
}

// ContainerIDForProject resolves the first devcontainer ID for a project path.
func ContainerIDForProject(projectDir string) string {
	out, err := runOutput(projectDir, "docker", "ps", "--filter", fmt.Sprintf("label=%s=%s", devcontainerFolderLabel, projectDir), "--format", "{{.ID}}")
	if err != nil {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		id := strings.TrimSpace(line)
		if id != "" {
			return id
		}
	}
	return ""
}

// Stop stops the given container ID.
func Stop(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("container ID is required")
	}
	if err := runCommand("", "docker", "stop", id); err != nil {
		return fmt.Errorf("stopping container %s: %w", id, err)
	}
	return nil
}

// Remove removes the given container ID.
func Remove(id string, removeVolumes bool) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("container ID is required")
	}
	args := []string{"rm"}
	if removeVolumes {
		args = append(args, "-v")
	}
	args = append(args, id)

	if err := runCommand("", "docker", args...); err != nil {
		return fmt.Errorf("removing container %s: %w", id, err)
	}
	return nil
}

// Prune removes stopped devcontainers and returns the count removed.
func Prune(removeVolumes bool) (int, error) {
	containers, err := List()
	if err != nil {
		return 0, err
	}

	removed := 0
	for _, c := range containers {
		if isRunningStatus(c.Status) {
			continue
		}
		if err := Remove(c.ID, removeVolumes); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

func isRunningStatus(status string) bool {
	return strings.HasPrefix(status, "Up ") || strings.HasPrefix(status, "Restarting")
}

func runCmd(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	return cmd.Run()
}

func runCmdOutput(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, string(out))
	}
	return string(out), nil
}
