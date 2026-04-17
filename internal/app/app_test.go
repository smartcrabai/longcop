package app

import (
	"context"
	"os/exec"
	"testing"

	"github.com/smartcrabai/longcop/internal/workflow"
)

type fakeUI struct {
	feature         string
	err             error
	yesNoAnswers    []bool
	yesNoQuestions  []string
	choiceAnswer    int
	choiceQuestions []string
}

func (f *fakeUI) ReadFeatureRequest() (string, error) {
	return f.feature, f.err
}

func (f *fakeUI) AskYesNo(question string, _ bool) (bool, error) {
	f.yesNoQuestions = append(f.yesNoQuestions, question)
	answer := f.yesNoAnswers[0]
	f.yesNoAnswers = f.yesNoAnswers[1:]
	return answer, nil
}

func (f *fakeUI) AskChoice(question string, _ []string) (int, error) {
	f.choiceQuestions = append(f.choiceQuestions, question)
	return f.choiceAnswer, nil
}

type fakeBootstrapper struct {
	ensureCalls int
	directories []string
}

func (f *fakeBootstrapper) Ensure(context.Context) error {
	f.ensureCalls++
	return nil
}

func (f *fakeBootstrapper) Directories() []string {
	return append([]string(nil), f.directories...)
}

type fakeLookup struct {
	paths map[string]string
	err   error
}

func (f fakeLookup) LookPath(file string) (string, error) {
	if path, ok := f.paths[file]; ok {
		return path, nil
	}
	if f.err != nil {
		return "", f.err
	}
	return "", exec.ErrNotFound
}

type fakeWorkspaceManager struct {
	workspace      workflow.Workspace
	prepareMode    workflow.ExecutionMode
	prepareRequest string
	prepareCalls   int
	cleanupCalls   int
	order          *[]string
}

func (f *fakeWorkspaceManager) Prepare(_ context.Context, mode workflow.ExecutionMode, featureRequest string) (workflow.Workspace, error) {
	f.prepareCalls++
	f.prepareMode = mode
	f.prepareRequest = featureRequest
	if f.order != nil {
		*f.order = append(*f.order, "prepare")
	}
	return f.workspace, nil
}

func (f *fakeWorkspaceManager) Cleanup(context.Context, workflow.Workspace) error {
	f.cleanupCalls++
	if f.order != nil {
		*f.order = append(*f.order, "cleanup")
	}
	return nil
}

type fakeRunner struct {
	result workflow.RunResult
	spec   workflow.RunSpec
	calls  int
	order  *[]string
}

func (f *fakeRunner) Run(_ context.Context, spec workflow.RunSpec) (workflow.RunResult, error) {
	f.calls++
	f.spec = spec
	if f.order != nil {
		*f.order = append(*f.order, "run")
	}
	return f.result, nil
}

type fakeMerger struct {
	repoRoot    string
	pullRequest workflow.PullRequest
	calls       int
	order       *[]string
}

func (f *fakeMerger) WaitAndMerge(_ context.Context, repoRoot string, pullRequest workflow.PullRequest) error {
	f.calls++
	f.repoRoot = repoRoot
	f.pullRequest = pullRequest
	if f.order != nil {
		*f.order = append(*f.order, "merge")
	}
	return nil
}

func TestRunUsesCurrentBranchModeWithoutPullRequest(t *testing.T) {
	t.Parallel()

	ui := &fakeUI{
		feature:      "Implement a new diff command",
		yesNoAnswers: []bool{true, false},
		choiceAnswer: 2,
	}
	bootstrapper := &fakeBootstrapper{directories: []string{"/skills", "/tools"}}
	workspaces := &fakeWorkspaceManager{
		workspace: workflow.Workspace{
			RepoRoot:      "/repo",
			WorkingDir:    "/repo",
			BaseBranch:    "main",
			CurrentBranch: "main",
		},
	}
	runner := &fakeRunner{}
	merger := &fakeMerger{}

	application := New(ui, bootstrapper, fakeLookup{paths: map[string]string{"coderabbit": "/usr/bin/coderabbit"}}, workspaces, runner, merger)

	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if bootstrapper.ensureCalls != 1 {
		t.Fatalf("expected skills bootstrap once, got %d", bootstrapper.ensureCalls)
	}
	if workspaces.prepareCalls != 1 {
		t.Fatalf("expected workspace preparation once, got %d", workspaces.prepareCalls)
	}
	if workspaces.cleanupCalls != 1 {
		t.Fatalf("expected workspace cleanup once, got %d", workspaces.cleanupCalls)
	}
	if runner.calls != 1 {
		t.Fatalf("expected Copilot runner once, got %d", runner.calls)
	}
	if merger.calls != 0 {
		t.Fatalf("expected no merge call, got %d", merger.calls)
	}
	if runner.spec.Options.Mode != workflow.ModeCurrentBranch {
		t.Fatalf("expected current branch mode, got %v", runner.spec.Options.Mode)
	}
	if runner.spec.Options.UseCodeRabbit {
		t.Fatal("expected coderabbit to be disabled")
	}
	if runner.spec.Options.AutoMerge {
		t.Fatal("expected auto-merge to stay disabled")
	}
	if len(ui.yesNoQuestions) != 2 {
		t.Fatalf("expected two yes/no questions, got %d", len(ui.yesNoQuestions))
	}
}

func TestRunCreatesPullRequestThenCleansUpBeforeAutoMerge(t *testing.T) {
	t.Parallel()

	order := make([]string, 0, 4)
	ui := &fakeUI{
		feature:      "Implement CI status summaries",
		yesNoAnswers: []bool{false, true},
		choiceAnswer: 0,
	}
	bootstrapper := &fakeBootstrapper{directories: []string{"/skills", "/tools"}}
	workspaces := &fakeWorkspaceManager{
		workspace: workflow.Workspace{
			RepoRoot:      "/repo",
			WorkingDir:    "/worktree",
			BaseBranch:    "main",
			CurrentBranch: "lcop/ci-status",
			UsesWorktree:  true,
			WorktreePath:  "/worktree",
		},
		order: &order,
	}
	runner := &fakeRunner{
		result: workflow.RunResult{
			PullRequest: &workflow.PullRequest{Number: 42, URL: "https://example.com/pull/42"},
		},
		order: &order,
	}
	merger := &fakeMerger{order: &order}

	application := New(ui, bootstrapper, fakeLookup{}, workspaces, runner, merger)

	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if merger.calls != 1 {
		t.Fatalf("expected auto-merge once, got %d", merger.calls)
	}
	if merger.pullRequest.Number != 42 {
		t.Fatalf("expected PR #42, got #%d", merger.pullRequest.Number)
	}
	if merger.repoRoot != "/repo" {
		t.Fatalf("expected auto-merge repo root /repo, got %s", merger.repoRoot)
	}
	expectedOrder := []string{"prepare", "run", "cleanup", "merge"}
	if len(order) != len(expectedOrder) {
		t.Fatalf("expected order %v, got %v", expectedOrder, order)
	}
	for index := range expectedOrder {
		if order[index] != expectedOrder[index] {
			t.Fatalf("expected order %v, got %v", expectedOrder, order)
		}
	}
}
