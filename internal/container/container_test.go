package container

import (
	"errors"
	"strings"
	"testing"
)

func TestListParsesDevcontainers(t *testing.T) {
	origRunOutput := runOutput
	defer func() { runOutput = origRunOutput }()

	runOutput = func(dir, name string, args ...string) (string, error) {
		return "abc123\tmy-project\tUp 2 hours\t2026-02-24 10:00:00 +0000 UTC\t/workspaces/my-project\n" +
			"def456\told-project\tExited (0) 1 day ago\t2026-02-20 08:00:00 +0000 UTC\t/workspaces/old-project\n", nil
	}

	containers, err := List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(containers))
	}
	if containers[0].ID != "abc123" || containers[0].ProjectDir != "/workspaces/my-project" {
		t.Fatalf("unexpected first container: %+v", containers[0])
	}
	if containers[1].Status != "Exited (0) 1 day ago" {
		t.Fatalf("unexpected status: %q", containers[1].Status)
	}
}

func TestContainerIDForProjectReturnsFirstMatch(t *testing.T) {
	origRunOutput := runOutput
	defer func() { runOutput = origRunOutput }()

	runOutput = func(dir, name string, args ...string) (string, error) {
		if dir != "/tmp/project" {
			t.Fatalf("expected project dir to be used, got %q", dir)
		}
		return "id-1\nid-2\n", nil
	}

	id := ContainerIDForProject("/tmp/project")
	if id != "id-1" {
		t.Fatalf("expected first ID, got %q", id)
	}
}

func TestStopAndRemoveInvokeDocker(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	var calls []string
	runCommand = func(dir, name string, args ...string) error {
		calls = append(calls, name+" "+strings.Join(args, " "))
		return nil
	}

	if err := Stop("abc"); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if err := Remove("def", true); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 docker calls, got %d", len(calls))
	}
	if calls[0] != "docker stop abc" {
		t.Fatalf("unexpected stop call: %q", calls[0])
	}
	if calls[1] != "docker rm -v def" {
		t.Fatalf("unexpected remove call: %q", calls[1])
	}
}

func TestPruneRemovesOnlyStoppedContainers(t *testing.T) {
	origRunOutput := runOutput
	origRunCommand := runCommand
	defer func() {
		runOutput = origRunOutput
		runCommand = origRunCommand
	}()

	runOutput = func(dir, name string, args ...string) (string, error) {
		return "run1\trunning\tUp 10 minutes\t2026-02-20\t/workspaces/run\n" +
			"stop1\tstopped\tExited (0) 2 hours ago\t2026-02-20\t/workspaces/stop\n", nil
	}

	var removedIDs []string
	runCommand = func(dir, name string, args ...string) error {
		if len(args) > 0 && args[0] == "rm" {
			removedIDs = append(removedIDs, args[len(args)-1])
		}
		return nil
	}

	count, err := Prune(false)
	if err != nil {
		t.Fatalf("Prune returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 removed container, got %d", count)
	}
	if len(removedIDs) != 1 || removedIDs[0] != "stop1" {
		t.Fatalf("unexpected removed IDs: %v", removedIDs)
	}
}

func TestListReturnsErrorWhenDockerFails(t *testing.T) {
	origRunOutput := runOutput
	defer func() { runOutput = origRunOutput }()

	runOutput = func(dir, name string, args ...string) (string, error) {
		return "", errors.New("docker unavailable")
	}

	_, err := List()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "listing devcontainers") {
		t.Fatalf("unexpected error: %v", err)
	}
}
