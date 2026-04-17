package copilot

import (
	"context"
	"testing"

	"longcop/internal/workflow"
)

type fakeShell struct {
	responses map[string]string
	calls     []string
}

func (f *fakeShell) Run(_ context.Context, dir string, name string, args ...string) (string, error) {
	key := dir + "::" + name
	for _, arg := range args {
		key += " " + arg
	}
	f.calls = append(f.calls, key)
	return f.responses[key], nil
}

func (f *fakeShell) LookPath(file string) (string, error) {
	return "/usr/bin/" + file, nil
}

func TestPullRequestToolCreatesAndRecordsPullRequest(t *testing.T) {
	t.Parallel()

	shell := &fakeShell{
		responses: map[string]string{
			"/worktree::git push --set-upstream origin lcop/feature":                                  "",
			"/worktree::gh pr create --base main --head lcop/feature --title Add feature --body Done": "https://github.com/example/repo/pull/13\n",
		},
	}
	tool := newPullRequestTool(context.Background(), shell, workflow.Workspace{
		WorkingDir:    "/worktree",
		BaseBranch:    "main",
		CurrentBranch: "lcop/feature",
	}, false)

	pullRequest, err := tool.create(context.Background(), pullRequestParams{
		Title: "Add feature",
		Body:  "Done",
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	if pullRequest.Number != 13 {
		t.Fatalf("expected PR #13, got #%d", pullRequest.Number)
	}
	if tool.result() == nil || tool.result().URL == "" {
		t.Fatal("expected tool to keep the created pull request result")
	}
}

func TestParsePullRequestOutputRejectsMissingURL(t *testing.T) {
	t.Parallel()

	if _, err := parsePullRequestOutput(" \n "); err == nil {
		t.Fatal("expected parsePullRequestOutput to reject empty output")
	}
}
