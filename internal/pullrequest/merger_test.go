package pullrequest

import (
	"context"
	"strings"
	"testing"

	"github.com/smartcrabai/longcop/internal/workflow"
)

type fakeRunner struct {
	calls []string
}

func (f *fakeRunner) Run(_ context.Context, dir string, name string, args ...string) (string, error) {
	f.calls = append(f.calls, strings.Join(append([]string{dir, name}, args...), " "))
	return "", nil
}

func (f *fakeRunner) LookPath(file string) (string, error) {
	return "", nil
}

func TestWaitAndMergeRunsChecksBeforeMerge(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	merger := NewAutoMerger(runner)

	if err := merger.WaitAndMerge(context.Background(), "/repo", workflow.PullRequest{Number: 21}); err != nil {
		t.Fatalf("WaitAndMerge returned error: %v", err)
	}

	expected := []string{
		"/repo gh pr checks 21 --watch",
		"/repo gh pr merge 21 --merge",
	}
	if len(runner.calls) != len(expected) {
		t.Fatalf("expected %d calls, got %v", len(expected), runner.calls)
	}
	for index := range expected {
		if runner.calls[index] != expected[index] {
			t.Fatalf("expected calls %v, got %v", expected, runner.calls)
		}
	}
}
