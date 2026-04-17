package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/smartcrabai/longcop/internal/config"
	"github.com/smartcrabai/longcop/internal/system"
	"github.com/smartcrabai/longcop/internal/workflow"
)

var invalidBranchCharacters = regexp.MustCompile(`[^a-z0-9]+`)

type Manager struct {
	runner      system.Runner
	homeDir     string
	originalDir string
	now         func() time.Time
}

func NewManager(runner system.Runner, homeDir string, originalDir string, now func() time.Time) *Manager {
	return &Manager{
		runner:      runner,
		homeDir:     homeDir,
		originalDir: originalDir,
		now:         now,
	}
}

func (m *Manager) Prepare(ctx context.Context, mode workflow.ExecutionMode, featureRequest string) (workflow.Workspace, error) {
	repoRoot, err := m.gitOutput(ctx, m.originalDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return workflow.Workspace{}, fmt.Errorf("resolve repository root: %w", err)
	}

	currentBranch, err := m.gitOutput(ctx, repoRoot, "branch", "--show-current")
	if err != nil {
		return workflow.Workspace{}, fmt.Errorf("resolve current branch: %w", err)
	}

	workspace := workflow.Workspace{
		RepoRoot:      repoRoot,
		WorkingDir:    repoRoot,
		BaseBranch:    currentBranch,
		CurrentBranch: currentBranch,
	}

	if !mode.RequiresWorktree() {
		return workspace, nil
	}

	root := filepath.Join(m.homeDir, config.WorktreeRootDirectoryName, config.WorktreeLeafDirectoryName)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return workflow.Workspace{}, fmt.Errorf("create worktree root %s: %w", root, err)
	}

	folderName := fmt.Sprintf("%s-%s", sanitizeFeatureSlug(featureRequest), m.now().UTC().Format("20060102150405"))
	branchName := fmt.Sprintf("%s/%s", config.AppName, folderName)
	worktreePath := filepath.Join(root, folderName)

	if _, err := m.runner.Run(ctx, repoRoot, "git", "worktree", "add", "-b", branchName, worktreePath, currentBranch); err != nil {
		return workflow.Workspace{}, fmt.Errorf("create worktree %s: %w", worktreePath, err)
	}

	workspace.WorkingDir = worktreePath
	workspace.CurrentBranch = branchName
	workspace.UsesWorktree = true
	workspace.WorktreePath = worktreePath

	return workspace, nil
}

func (m *Manager) Cleanup(ctx context.Context, workspace workflow.Workspace) error {
	if !workspace.UsesWorktree {
		return nil
	}

	if _, err := m.runner.Run(ctx, workspace.RepoRoot, "git", "worktree", "remove", "--force", workspace.WorktreePath); err != nil {
		return fmt.Errorf("remove worktree %s: %w", workspace.WorktreePath, err)
	}

	return nil
}

func (m *Manager) gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	output, err := m.runner.Run(ctx, dir, "git", args...)
	if err != nil {
		return "", err
	}

	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return "", fmt.Errorf("git %s returned empty output", strings.Join(args, " "))
	}

	return trimmed, nil
}

func sanitizeFeatureSlug(featureRequest string) string {
	cleaned := strings.ToLower(featureRequest)
	cleaned = invalidBranchCharacters.ReplaceAllString(cleaned, "-")
	cleaned = strings.Trim(cleaned, "-")
	if cleaned == "" {
		return config.DefaultFeatureSlug
	}
	if len(cleaned) <= config.MaxFeatureSlugSize {
		return cleaned
	}
	cleaned = strings.Trim(cleaned[:config.MaxFeatureSlugSize], "-")
	if cleaned == "" {
		return config.DefaultFeatureSlug
	}
	return cleaned
}
