package workspace

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/smartcrabai/longcop/internal/system"
	"github.com/smartcrabai/longcop/internal/workflow"
)

type fakeRunner struct {
	responses map[string]string
	calls     []string
}

func (f *fakeRunner) Run(_ context.Context, dir string, name string, args ...string) (string, error) {
	command := strings.Join(append([]string{name}, args...), " ")
	f.calls = append(f.calls, dir+"::"+command)
	return f.responses[dir+"::"+command], nil
}

func (f *fakeRunner) LookPath(file string) (string, error) {
	return "", nil
}

var _ system.Runner = (*fakeRunner)(nil)

func TestPrepareReturnsRepositoryWorkspaceForCurrentBranchMode(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{
		responses: map[string]string{
			"/repo::git rev-parse --show-toplevel": "/repo\n",
			"/repo::git branch --show-current":     "main\n",
		},
	}

	manager := NewManager(runner, t.TempDir(), "/repo", time.Now)
	workspace, err := manager.Prepare(context.Background(), workflow.ModeCurrentBranch, "Add search")
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	if workspace.WorkingDir != "/repo" {
		t.Fatalf("expected working dir /repo, got %s", workspace.WorkingDir)
	}
	if workspace.UsesWorktree {
		t.Fatal("expected current-branch mode to skip worktree creation")
	}
	if len(runner.calls) != 2 {
		t.Fatalf("expected two git calls, got %v", runner.calls)
	}
}

func TestPrepareCreatesWorktreeWithSanitizedBranchName(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	runner := &fakeRunner{
		responses: map[string]string{
			"/repo::git rev-parse --show-toplevel": "/repo\n",
			"/repo::git branch --show-current":     "main\n",
		},
	}

	now := func() time.Time {
		return time.Date(2026, time.April, 17, 18, 30, 0, 0, time.UTC)
	}

	manager := NewManager(runner, homeDir, "/repo", now)
	workspace, err := manager.Prepare(context.Background(), workflow.ModeWorktreePullRequest, "Add release summary output!!!")
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	expectedFolder := "add-release-summary-output-20260417183000"
	expectedPath := filepath.Join(homeDir, ".longcop", "worktrees", expectedFolder)
	if workspace.WorktreePath != expectedPath {
		t.Fatalf("expected worktree path %s, got %s", expectedPath, workspace.WorktreePath)
	}
	if workspace.CurrentBranch != "lcop/"+expectedFolder {
		t.Fatalf("unexpected branch name: %s", workspace.CurrentBranch)
	}
	if !workspace.UsesWorktree {
		t.Fatal("expected worktree usage to be enabled")
	}

	lastCall := runner.calls[len(runner.calls)-1]
	if !strings.Contains(lastCall, "git worktree add -b lcop/"+expectedFolder+" "+expectedPath+" main") {
		t.Fatalf("expected worktree creation call, got %s", lastCall)
	}
}

func TestCleanupRemovesWorktreeWhenNeeded(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: map[string]string{}}
	manager := NewManager(runner, t.TempDir(), "/repo", time.Now)

	err := manager.Cleanup(context.Background(), workflow.Workspace{
		RepoRoot:     "/repo",
		UsesWorktree: true,
		WorktreePath: "/worktree",
	})
	if err != nil {
		t.Fatalf("Cleanup returned error: %v", err)
	}

	if len(runner.calls) != 1 || runner.calls[0] != "/repo::git worktree remove --force /worktree" {
		t.Fatalf("unexpected cleanup calls: %v", runner.calls)
	}
}
