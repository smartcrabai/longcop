package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"longcop/internal/app"
	"longcop/internal/console"
	copilotrunner "longcop/internal/copilot"
	"longcop/internal/pullrequest"
	"longcop/internal/skills"
	"longcop/internal/system"
	"longcop/internal/workspace"
)

func main() {
	ctx := context.Background()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		exitWithError(fmt.Errorf("resolve home directory: %w", err))
	}

	workingDir, err := os.Getwd()
	if err != nil {
		exitWithError(fmt.Errorf("resolve working directory: %w", err))
	}

	terminal, err := console.Open(os.Stdin, os.Stdout)
	if err != nil {
		exitWithError(fmt.Errorf("open terminal: %w", err))
	}
	defer func() {
		if err := terminal.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	shell := system.ShellRunner{}
	application := app.New(
		terminal,
		skills.NewBootstrapper(homeDir),
		shell,
		workspace.NewManager(shell, homeDir, workingDir, time.Now),
		copilotrunner.NewRunner(shell, os.Stdout),
		pullrequest.NewAutoMerger(shell),
	)

	if err := application.Run(ctx); err != nil {
		exitWithError(err)
	}
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
