package workflow

import "fmt"

type ExecutionMode int

const (
	ModeWorktreePullRequest ExecutionMode = iota
	ModeWorktreeDraftPullRequest
	ModeCurrentBranch
)

var executionModeLabels = []string{
	"Use Git worktree and create a pull request",
	"Use Git worktree and create a draft pull request",
	"Implement on the current branch without a worktree",
}

func ModeLabels() []string {
	return append([]string(nil), executionModeLabels...)
}

func ModeFromIndex(index int) (ExecutionMode, error) {
	switch index {
	case 0:
		return ModeWorktreePullRequest, nil
	case 1:
		return ModeWorktreeDraftPullRequest, nil
	case 2:
		return ModeCurrentBranch, nil
	default:
		return ModeCurrentBranch, fmt.Errorf("unsupported execution mode index %d", index)
	}
}

func (m ExecutionMode) RequiresWorktree() bool {
	return m != ModeCurrentBranch
}

func (m ExecutionMode) RequiresPullRequest() bool {
	return m != ModeCurrentBranch
}

func (m ExecutionMode) DraftPullRequest() bool {
	return m == ModeWorktreeDraftPullRequest
}

type Options struct {
	FeatureRequest string
	UseTDD         bool
	Mode           ExecutionMode
	UseCodeRabbit  bool
	AutoMerge      bool
}

type Workspace struct {
	RepoRoot      string
	WorkingDir    string
	BaseBranch    string
	CurrentBranch string
	UsesWorktree  bool
	WorktreePath  string
}

type PullRequest struct {
	Number int
	URL    string
}

type RunSpec struct {
	Options          Options
	Workspace        Workspace
	SkillDirectories []string
}

type RunResult struct {
	PullRequest *PullRequest
}
