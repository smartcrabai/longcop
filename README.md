# longcop

A CLI tool that drives [GitHub Copilot CLI](https://github.com/github/copilot-sdk) through a structured implementation workflow with optional TDD, code review, CI debugging, and auto-merge.

## How it works

`lcop` collects a feature request from the user, asks a few questions, and then delegates the entire implementation to GitHub Copilot SDK:

1. **Skill bootstrapping** — Creates skill definitions under `~/.lcop/` on startup.
2. **Interactive prompt** — Accepts a feature request, then asks:
   - TDD or standard implementation
   - Execution mode: worktree + PR, worktree + draft PR, or current branch
   - Optional coderabbit review (if `coderabbit` CLI is available)
   - Auto-merge after CI passes
3. **Copilot workflow** — Runs a structured Copilot session that implements the feature, simplifies the code, checks for AI anti-patterns, optionally runs coderabbit, creates a PR, and monitors CI.

## Installation

Download the latest release from [GitHub Releases](https://github.com/smartcrabai/longcop/releases).

Or build from source:

```bash
go build -o lcop ./cmd/lcop
```

## Usage

```bash
lcop
```

The tool is interactive — it will guide you through the workflow. Press `Ctrl+D` to exit.

## Built-in skills

| Skill | Location | Purpose |
|-------|----------|---------|
| `lcop-simplify` | `~/.lcop/skills/simplify/` | Reduces duplication and removes dead code after implementation |
| `lcop-ai-anti-pattern` | `~/.lcop/skills/ai-anti-pattern/` | Detects AI-assisted coding anti-patterns |
| `lcop-ci-debugger` | `~/.lcop/skills/ci-debugger/` | Investigates CI failures and drives the branch to green |
| `lcop-coderabbit` | `~/.lcop/tools/coderabbit/` | Runs coderabbit review and applies fixes |

## Requirements

- Go 1.26+
- [GitHub Copilot CLI](https://github.com/github/copilot-sdk)
- Git
- (Optional) [coderabbit](https://coderabbit.ai/) CLI for automated code review
