package prompt

import (
	"fmt"
	"strings"

	"github.com/smartcrabai/longcop/internal/config"
	"github.com/smartcrabai/longcop/internal/workflow"
)

type Builder struct{}

func NewBuilder() Builder {
	return Builder{}
}

func (Builder) Build(spec workflow.RunSpec) string {
	steps := make([]string, 0, 10)

	if spec.Options.UseTDD {
		steps = append(steps, "Write or update tests before implementation.")
	} else {
		steps = append(steps, "Implement the requested behavior directly and add tests when they are needed to validate behavior.")
	}

	steps = append(
		steps,
		"Implement the requested behavior.",
		"Run tests and lint whenever they help you validate progress.",
		"If anything is still missing, return to the earliest relevant step and continue until the request is complete.",
		fmt.Sprintf("Use the %s skill to simplify the implementation and remove duplication.", config.SimplifySkillName),
		fmt.Sprintf("Use the %s skill to review the new code for AI-style anti-patterns and fix any real issues. That skill delegates to the globally installed %s skill.", config.AIAntiPatternSkillName, config.GlobalAIAntipatternSkillName),
	)

	if spec.Options.UseCodeRabbit {
		steps = append(steps, fmt.Sprintf("Use the %s skill to run coderabbit review and apply worthwhile fixes.", config.CodeRabbitSkillName))
	}

	steps = append(steps, "Run the relevant tests and lint again, then fix any remaining failures.")

	if spec.Options.Mode.RequiresPullRequest() {
		prKind := "pull request"
		if spec.Options.Mode.DraftPullRequest() {
			prKind = "draft pull request"
		}

		steps = append(
			steps,
			fmt.Sprintf("Use the %s tool to create a %s for branch %s against base %s.", config.PullRequestCreatorToolName, prKind, spec.Workspace.CurrentBranch, spec.Workspace.BaseBranch),
			fmt.Sprintf("Use the %s skill to monitor CI for that pull request and fix any failures until it is green.", config.CIDebuggerSkillName),
		)
	}

	lines := []string{
		"The user asked for the following feature to be implemented:",
		spec.Options.FeatureRequest,
		"",
		"Implementation constraints:",
		"- All implementation details, prompts, comments, and identifiers must be written in English.",
		fmt.Sprintf("- Work only inside %s.", spec.Workspace.WorkingDir),
		"- Do not switch branches manually; the host application already prepared the workspace.",
	}

	if spec.Options.Mode.RequiresPullRequest() {
		lines = append(lines, "- The host application will handle any requested auto-merge after the pull request is green.")
	}

	lines = append(lines, "", "Follow this workflow:")
	for index, step := range steps {
		lines = append(lines, fmt.Sprintf("%d. %s", index+1, step))
	}

	return strings.Join(lines, "\n")
}
