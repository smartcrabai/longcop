package copilot

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	copilotsdk "github.com/github/copilot-sdk/go"

	"longcop/internal/config"
	"longcop/internal/errutil"
	"longcop/internal/prompt"
	"longcop/internal/system"
	"longcop/internal/workflow"
)

type Runner struct {
	shell   system.Runner
	out     io.Writer
	builder prompt.Builder
}

type pullRequestTool struct {
	ctx       context.Context
	shell     system.Runner
	workspace workflow.Workspace
	draft     bool
	created   *workflow.PullRequest
}

type pullRequestParams struct {
	Title string `json:"title" jsonschema:"Pull request title"`
	Body  string `json:"body" jsonschema:"Pull request body"`
}

func NewRunner(shell system.Runner, out io.Writer) *Runner {
	return &Runner{
		shell:   shell,
		out:     out,
		builder: prompt.NewBuilder(),
	}
}

func (r *Runner) Run(ctx context.Context, spec workflow.RunSpec) (result workflow.RunResult, err error) {
	cliPath, err := r.resolveCLIPath()
	if err != nil {
		return workflow.RunResult{}, err
	}

	clientOptions := &copilotsdk.ClientOptions{CLIPath: cliPath}
	client := copilotsdk.NewClient(clientOptions)
	if err := client.Start(ctx); err != nil {
		return workflow.RunResult{}, fmt.Errorf("start Copilot SDK client: %w", err)
	}
	defer func() {
		err = errors.Join(err, errutil.Wrap("stop Copilot SDK client", client.Stop()))
	}()

	sessionConfig := &copilotsdk.SessionConfig{
		ClientName:            config.CopilotClientName,
		EnableConfigDiscovery: true,
		OnPermissionRequest:   copilotsdk.PermissionHandler.ApproveAll,
		SkillDirectories:      spec.SkillDirectories,
		WorkingDirectory:      spec.Workspace.WorkingDir,
	}

	prTool := newPullRequestTool(ctx, r.shell, spec.Workspace, spec.Options.Mode.DraftPullRequest())
	if spec.Options.Mode.RequiresPullRequest() {
		sessionConfig.Tools = []copilotsdk.Tool{prTool.definition()}
	}

	session, err := client.CreateSession(ctx, sessionConfig)
	if err != nil {
		return workflow.RunResult{}, fmt.Errorf("create Copilot SDK session: %w", err)
	}
	defer func() {
		err = errors.Join(err, errutil.Wrap("disconnect Copilot SDK session", session.Disconnect()))
	}()

	reply, err := session.SendAndWait(ctx, copilotsdk.MessageOptions{Prompt: r.builder.Build(spec)})
	if err != nil {
		return workflow.RunResult{}, fmt.Errorf("run Copilot SDK prompt: %w", err)
	}

	if r.out != nil && reply != nil {
		if message, ok := reply.Data.(*copilotsdk.AssistantMessageData); ok && strings.TrimSpace(message.Content) != "" {
			if _, err := fmt.Fprintln(r.out, message.Content); err != nil {
				return workflow.RunResult{}, fmt.Errorf("write assistant output: %w", err)
			}
		}
	}

	result = workflow.RunResult{PullRequest: prTool.result()}
	return result, nil
}

func (r *Runner) resolveCLIPath() (string, error) {
	if cliPath := strings.TrimSpace(os.Getenv(config.CopilotCLIPathEnv)); cliPath != "" {
		return cliPath, nil
	}

	cliPath, err := r.shell.LookPath("copilot")
	if err == nil {
		return cliPath, nil
	}

	return "", fmt.Errorf("copilot CLI was not found; install `copilot` or set %s", config.CopilotCLIPathEnv)
}

func newPullRequestTool(ctx context.Context, shell system.Runner, workspace workflow.Workspace, draft bool) *pullRequestTool {
	return &pullRequestTool{
		ctx:       ctx,
		shell:     shell,
		workspace: workspace,
		draft:     draft,
	}
}

func (t *pullRequestTool) definition() copilotsdk.Tool {
	return copilotsdk.DefineTool(
		config.PullRequestCreatorToolName,
		"Creates a GitHub pull request for the current branch.",
		func(params pullRequestParams, _ copilotsdk.ToolInvocation) (workflow.PullRequest, error) {
			return t.create(t.ctx, params)
		},
	)
}

func (t *pullRequestTool) create(ctx context.Context, params pullRequestParams) (workflow.PullRequest, error) {
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return workflow.PullRequest{}, errors.New("pull request title cannot be empty")
	}

	body := strings.TrimSpace(params.Body)
	if body == "" {
		return workflow.PullRequest{}, errors.New("pull request body cannot be empty")
	}

	if _, err := t.shell.Run(ctx, t.workspace.WorkingDir, "git", "push", "--set-upstream", "origin", t.workspace.CurrentBranch); err != nil {
		return workflow.PullRequest{}, fmt.Errorf("push implementation branch: %w", err)
	}

	args := []string{
		"pr",
		"create",
		"--base", t.workspace.BaseBranch,
		"--head", t.workspace.CurrentBranch,
		"--title", title,
		"--body", body,
	}
	if t.draft {
		args = append(args, "--draft")
	}

	output, err := t.shell.Run(ctx, t.workspace.WorkingDir, "gh", args...)
	if err != nil {
		return workflow.PullRequest{}, fmt.Errorf("create pull request: %w", err)
	}

	pullRequest, err := parsePullRequestOutput(output)
	if err != nil {
		return workflow.PullRequest{}, err
	}

	t.created = &pullRequest
	return pullRequest, nil
}

func (t *pullRequestTool) result() *workflow.PullRequest {
	if t.created == nil {
		return nil
	}

	pullRequest := *t.created
	return &pullRequest
}

func parsePullRequestOutput(output string) (workflow.PullRequest, error) {
	var rawURL string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		for _, field := range strings.Fields(line) {
			trimmed := strings.TrimSpace(field)
			if strings.HasPrefix(trimmed, "https://") || strings.HasPrefix(trimmed, "http://") {
				rawURL = trimmed
				break
			}
		}
		if rawURL != "" {
			break
		}
	}

	if rawURL == "" {
		return workflow.PullRequest{}, errors.New("pull request creation output did not contain a URL")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return workflow.PullRequest{}, fmt.Errorf("parse pull request URL %q: %w", rawURL, err)
	}

	number, err := strconv.Atoi(path.Base(parsedURL.Path))
	if err != nil {
		return workflow.PullRequest{}, fmt.Errorf("parse pull request number from %q: %w", rawURL, err)
	}

	return workflow.PullRequest{
		Number: number,
		URL:    rawURL,
	}, nil
}
