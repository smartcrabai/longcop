package prompt

import (
	"strings"
	"testing"

	"github.com/smartcrabai/longcop/internal/config"
	"github.com/smartcrabai/longcop/internal/workflow"
)

func TestBuildIncludesSimplifyAndAntiPatternSteps(t *testing.T) {
	t.Parallel()

	prompt := NewBuilder().Build(workflow.RunSpec{
		Options: workflow.Options{
			FeatureRequest: "Implement release notes generation",
			UseTDD:         true,
			Mode:           workflow.ModeCurrentBranch,
			UseCodeRabbit:  true,
		},
		Workspace: workflow.Workspace{
			WorkingDir:    "/repo",
			BaseBranch:    "main",
			CurrentBranch: "main",
		},
	})

	for _, needle := range []string{
		"Write or update tests before implementation.",
		config.SimplifySkillName,
		config.AIAntiPatternSkillName,
		config.GlobalAIAntipatternSkillName,
		config.CodeRabbitSkillName,
	} {
		if !strings.Contains(prompt, needle) {
			t.Fatalf("expected prompt to contain %q, got %s", needle, prompt)
		}
	}

	if strings.Contains(prompt, config.PullRequestCreatorToolName) {
		t.Fatalf("did not expect prompt to mention %s in current-branch mode", config.PullRequestCreatorToolName)
	}
}

func TestBuildIncludesPullRequestWorkflowWhenRequested(t *testing.T) {
	t.Parallel()

	prompt := NewBuilder().Build(workflow.RunSpec{
		Options: workflow.Options{
			FeatureRequest: "Implement draft release workflow",
			UseTDD:         false,
			Mode:           workflow.ModeWorktreeDraftPullRequest,
		},
		Workspace: workflow.Workspace{
			WorkingDir:    "/worktree",
			BaseBranch:    "main",
			CurrentBranch: "lcop/draft-release",
		},
	})

	for _, needle := range []string{
		config.PullRequestCreatorToolName,
		"draft pull request",
		config.CIDebuggerSkillName,
	} {
		if !strings.Contains(prompt, needle) {
			t.Fatalf("expected prompt to contain %q, got %s", needle, prompt)
		}
	}
}
