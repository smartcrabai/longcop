package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/smartcrabai/longcop/internal/app"
	"github.com/smartcrabai/longcop/internal/console"
	copilotrunner "github.com/smartcrabai/longcop/internal/copilot"
	"github.com/smartcrabai/longcop/internal/pullrequest"
	"github.com/smartcrabai/longcop/internal/skills"
	"github.com/smartcrabai/longcop/internal/system"
	"github.com/smartcrabai/longcop/internal/workspace"
)

const devVersion = "dev"

var version = devVersion

const helpText = `lcop — CLI that drives GitHub Copilot CLI through a structured implementation workflow.

Usage:
  lcop [flags]

Flags:
  -h, --help     Show this help and exit
  -v, --version  Show version and exit

lcop is interactive. It prompts for a feature request and then asks a few
configuration questions (TDD, execution mode, coderabbit, auto-merge).
`

func main() {
	if args := os.Args[1:]; len(args) > 0 {
		if len(args) == 1 {
			switch args[0] {
			case "-h", "--help":
				fmt.Fprint(os.Stdout, helpText)
				return
			case "-v", "--version":
				fmt.Println(resolveVersion())
				return
			}
		}
		fmt.Fprintf(os.Stderr, "lcop: unexpected arguments %v\n\n", args)
		fmt.Fprint(os.Stderr, helpText)
		os.Exit(2)
	}

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

func resolveVersion() string {
	if version != devVersion {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return devVersion
}
