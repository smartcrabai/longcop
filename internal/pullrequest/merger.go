package pullrequest

import (
	"context"
	"fmt"
	"strconv"

	"longcop/internal/system"
	"longcop/internal/workflow"
)

type AutoMerger struct {
	runner system.Runner
}

func NewAutoMerger(runner system.Runner) *AutoMerger {
	return &AutoMerger{runner: runner}
}

func (m *AutoMerger) WaitAndMerge(ctx context.Context, repoRoot string, pullRequest workflow.PullRequest) error {
	number := strconv.Itoa(pullRequest.Number)

	if _, err := m.runner.Run(ctx, repoRoot, "gh", "pr", "checks", number, "--watch"); err != nil {
		return fmt.Errorf("wait for CI on PR #%d: %w", pullRequest.Number, err)
	}

	if _, err := m.runner.Run(ctx, repoRoot, "gh", "pr", "merge", number, "--merge"); err != nil {
		return fmt.Errorf("merge PR #%d: %w", pullRequest.Number, err)
	}

	return nil
}
