package app

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/smartcrabai/longcop/internal/errutil"
	"github.com/smartcrabai/longcop/internal/workflow"
)

type UI interface {
	ReadFeatureRequest() (string, error)
	AskYesNo(question string, defaultYes bool) (bool, error)
	AskChoice(question string, options []string) (int, error)
}

type SkillBootstrapper interface {
	Ensure(ctx context.Context) error
	Directories() []string
}

type PathLookup interface {
	LookPath(file string) (string, error)
}

type WorkspaceManager interface {
	Prepare(ctx context.Context, mode workflow.ExecutionMode, featureRequest string) (workflow.Workspace, error)
	Cleanup(ctx context.Context, workspace workflow.Workspace) error
}

type CopilotRunner interface {
	Run(ctx context.Context, spec workflow.RunSpec) (workflow.RunResult, error)
}

type AutoMerger interface {
	WaitAndMerge(ctx context.Context, repoRoot string, pullRequest workflow.PullRequest) error
}

type App struct {
	ui         UI
	skills     SkillBootstrapper
	lookup     PathLookup
	workspaces WorkspaceManager
	runner     CopilotRunner
	merger     AutoMerger
}

func New(
	ui UI,
	skills SkillBootstrapper,
	lookup PathLookup,
	workspaces WorkspaceManager,
	runner CopilotRunner,
	merger AutoMerger,
) *App {
	return &App{
		ui:         ui,
		skills:     skills,
		lookup:     lookup,
		workspaces: workspaces,
		runner:     runner,
		merger:     merger,
	}
}

func (a *App) Run(ctx context.Context) error {
	if err := a.skills.Ensure(ctx); err != nil {
		return fmt.Errorf("bootstrap skills: %w", err)
	}

	featureRequest, err := a.ui.ReadFeatureRequest()
	if err != nil {
		return fmt.Errorf("collect feature request: %w", err)
	}

	featureRequest = strings.TrimSpace(featureRequest)
	if featureRequest == "" {
		return errors.New("feature request cannot be empty")
	}

	options, err := a.collectOptions(featureRequest)
	if err != nil {
		return err
	}

	workspace, err := a.workspaces.Prepare(ctx, options.Mode, featureRequest)
	if err != nil {
		return fmt.Errorf("prepare workspace: %w", err)
	}

	result, runErr := a.runner.Run(ctx, workflow.RunSpec{
		Options:          options,
		Workspace:        workspace,
		SkillDirectories: a.skills.Directories(),
	})
	cleanupErr := a.workspaces.Cleanup(ctx, workspace)

	if runErr != nil || cleanupErr != nil {
		return errors.Join(
			errutil.Wrap("run Copilot workflow", runErr),
			errutil.Wrap("cleanup workspace", cleanupErr),
		)
	}

	if options.Mode.RequiresPullRequest() && result.PullRequest == nil {
		return errors.New("workflow completed without creating the requested pull request")
	}

	if options.AutoMerge {
		if result.PullRequest == nil {
			return errors.New("auto-merge was requested but no pull request was created")
		}
		if err := a.merger.WaitAndMerge(ctx, workspace.RepoRoot, *result.PullRequest); err != nil {
			return fmt.Errorf("auto-merge pull request: %w", err)
		}
	}

	return nil
}

func (a *App) collectOptions(featureRequest string) (workflow.Options, error) {
	useTDD, err := a.ui.AskYesNo("Implement with TDD?", true)
	if err != nil {
		return workflow.Options{}, fmt.Errorf("ask TDD preference: %w", err)
	}

	modeIndex, err := a.ui.AskChoice("Select the implementation mode:", workflow.ModeLabels())
	if err != nil {
		return workflow.Options{}, fmt.Errorf("ask implementation mode: %w", err)
	}

	mode, err := workflow.ModeFromIndex(modeIndex)
	if err != nil {
		return workflow.Options{}, err
	}

	useCodeRabbit := false
	if _, err := a.lookup.LookPath("coderabbit"); err == nil {
		useCodeRabbit, err = a.ui.AskYesNo("Use coderabbit for review and fixes?", true)
		if err != nil {
			return workflow.Options{}, fmt.Errorf("ask coderabbit preference: %w", err)
		}
	} else if !errors.Is(err, exec.ErrNotFound) {
		return workflow.Options{}, fmt.Errorf("detect coderabbit command: %w", err)
	}

	autoMerge := false
	if mode.RequiresPullRequest() {
		autoMerge, err = a.ui.AskYesNo("Merge the pull request after CI passes?", true)
		if err != nil {
			return workflow.Options{}, fmt.Errorf("ask auto-merge preference: %w", err)
		}
	}

	return workflow.Options{
		FeatureRequest: featureRequest,
		UseTDD:         useTDD,
		Mode:           mode,
		UseCodeRabbit:  useCodeRabbit,
		AutoMerge:      autoMerge,
	}, nil
}
