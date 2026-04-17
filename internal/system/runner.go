package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(ctx context.Context, dir string, name string, args ...string) (string, error)
	LookPath(file string) (string, error)
}

type ShellRunner struct{}

func (ShellRunner) Run(ctx context.Context, dir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		trimmedOutput := strings.TrimSpace(string(output))
		commandText := strings.Join(append([]string{name}, args...), " ")
		if trimmedOutput == "" {
			return string(output), fmt.Errorf("run %q: %w", commandText, err)
		}
		return string(output), fmt.Errorf("run %q: %w: %s", commandText, err, trimmedOutput)
	}

	return string(output), nil
}

func (ShellRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}
