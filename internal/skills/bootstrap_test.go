package skills

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/smartcrabai/longcop/internal/config"
)

func TestEnsureCreatesExpectedSkillFiles(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	bootstrapper := NewBootstrapper(homeDir)

	if err := bootstrapper.Ensure(context.Background()); err != nil {
		t.Fatalf("Ensure returned error: %v", err)
	}

	checks := map[string]string{
		filepath.Join(homeDir, config.ConfigDirectoryName, config.SkillsDirectoryName, config.SimplifySkillDirectoryName, config.SkillFileName):      config.SimplifySkillName,
		filepath.Join(homeDir, config.ConfigDirectoryName, config.SkillsDirectoryName, config.AIAntiPatternSkillDirectoryName, config.SkillFileName): config.GlobalAIAntipatternSkillName,
		filepath.Join(homeDir, config.ConfigDirectoryName, config.SkillsDirectoryName, config.CIDebuggerSkillDirectoryName, config.SkillFileName):    config.CIDebuggerSkillName,
		filepath.Join(homeDir, config.ConfigDirectoryName, config.ToolsDirectoryName, config.CodeRabbitToolDirectoryName, config.SkillFileName):      config.CodeRabbitSkillName,
	}

	for filePath, needle := range checks {
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("read %s: %v", filePath, err)
		}
		if !strings.Contains(string(content), needle) {
			t.Fatalf("expected %s to contain %q, got %s", filePath, needle, string(content))
		}
	}
}
