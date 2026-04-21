package config

import "time"

const (
	AppName                         = "lcop"
	ConfigDirectoryName             = ".lcop"
	SkillsDirectoryName             = "skills"
	ToolsDirectoryName              = "tools"
	SkillFileName                   = "SKILL.md"
	SimplifySkillDirectoryName      = "simplify"
	AIAntiPatternSkillDirectoryName = "ai-anti-pattern"
	CIDebuggerSkillDirectoryName    = "ci-debugger"
	CodeRabbitToolDirectoryName     = "coderabbit"

	SimplifySkillName            = "lcop-simplify"
	AIAntiPatternSkillName       = "lcop-ai-anti-pattern"
	GlobalAIAntipatternSkillName = "ai-antipattern"
	CIDebuggerSkillName          = "lcop-ci-debugger"
	CodeRabbitSkillName          = "lcop-coderabbit"

	PullRequestCreatorToolName = "PullRequestCreator"
	CopilotClientName          = "longcop"
	CopilotCLIPathEnv          = "COPILOT_CLI_PATH"
	CopilotReplyTimeout        = 196 * time.Hour

	WorktreeRootDirectoryName = ".longcop"
	WorktreeLeafDirectoryName = "worktrees"

	DefaultFeatureSlug = "change"
	MaxFeatureSlugSize = 32
)
