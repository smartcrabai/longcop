package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"longcop/internal/config"
)

type Bootstrapper struct {
	homeDir string
}

type fileSpec struct {
	path    string
	content string
}

func NewBootstrapper(homeDir string) *Bootstrapper {
	return &Bootstrapper{homeDir: homeDir}
}

func (b *Bootstrapper) Ensure(ctx context.Context) error {
	for _, spec := range b.fileSpecs() {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(spec.path), 0o755); err != nil {
			return fmt.Errorf("create skill directory for %s: %w", spec.path, err)
		}
		existing, err := os.ReadFile(spec.path)
		if err == nil && string(existing) == spec.content {
			continue
		}
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read skill file %s: %w", spec.path, err)
		}
		if err := os.WriteFile(spec.path, []byte(spec.content), 0o644); err != nil {
			return fmt.Errorf("write skill file %s: %w", spec.path, err)
		}
	}

	return nil
}

func (b *Bootstrapper) Directories() []string {
	return []string{b.skillsRoot(), b.toolsRoot()}
}

func (b *Bootstrapper) fileSpecs() []fileSpec {
	return []fileSpec{
		{
			path: filepath.Join(
				b.skillsRoot(),
				config.SimplifySkillDirectoryName,
				config.SkillFileName,
			),
			content: simplifySkillContent(),
		},
		{
			path: filepath.Join(
				b.skillsRoot(),
				config.AIAntiPatternSkillDirectoryName,
				config.SkillFileName,
			),
			content: aiAntiPatternCompatibilityContent(),
		},
		{
			path: filepath.Join(
				b.skillsRoot(),
				config.CIDebuggerSkillDirectoryName,
				config.SkillFileName,
			),
			content: ciDebuggerSkillContent(),
		},
		{
			path: filepath.Join(
				b.toolsRoot(),
				config.CodeRabbitToolDirectoryName,
				config.SkillFileName,
			),
			content: codeRabbitSkillContent(),
		},
	}
}

func (b *Bootstrapper) skillsRoot() string {
	return filepath.Join(b.homeDir, config.ConfigDirectoryName, config.SkillsDirectoryName)
}

func (b *Bootstrapper) toolsRoot() string {
	return filepath.Join(b.homeDir, config.ConfigDirectoryName, config.ToolsDirectoryName)
}

func simplifySkillContent() string {
	return fmt.Sprintf(`---
name: %s
description: Simplify freshly implemented code without changing behavior.
---

Use this skill after implementation work is complete.
1. Reduce duplication by extracting the smallest useful helpers.
2. Remove unnecessary branching, temporary state, and dead code.
3. Preserve behavior, tests, and public interfaces unless the user asked for a change.
4. Keep code, comments, prompts, and identifiers in English.
`, config.SimplifySkillName)
}

func aiAntiPatternCompatibilityContent() string {
	return fmt.Sprintf(`---
name: %s
description: Compatibility wrapper for the globally installed %s skill.
---

Use the globally installed %s skill to review the newly changed code for AI-assisted anti-patterns.
Focus on misleading abstractions, duplicated logic, dead helpers, and unnecessary conditionals.
Apply only concrete fixes that improve maintainability and re-run validation after making changes.
`, config.AIAntiPatternSkillName, config.GlobalAIAntipatternSkillName, config.GlobalAIAntipatternSkillName)
}

func ciDebuggerSkillContent() string {
	return fmt.Sprintf(`---
name: %s
description: Investigate pull request CI failures and drive the branch back to green.
---

Use this skill after a pull request has been created.
1. Inspect the CI status for the current pull request.
2. Read the failing logs and identify the real root cause.
3. Apply the smallest fix that resolves the failure.
4. Re-run validation and repeat until the pull request is green or an external blocker remains.
`, config.CIDebuggerSkillName)
}

func codeRabbitSkillContent() string {
	return fmt.Sprintf(`---
name: %s
description: Use the coderabbit CLI review flow and apply worthwhile fixes.
---

Use this skill only when the coderabbit command is available and the user opted in.
1. Run coderabbit against the current changes.
2. Apply concrete, high-signal fixes.
3. Re-run the relevant validation.
`, config.CodeRabbitSkillName)
}
