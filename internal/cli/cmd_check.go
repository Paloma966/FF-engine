package cli

import (
	"fmt"
	"os"

	"github.com/chun/fiction_factory/internal/check"
	"github.com/chun/fiction_factory/internal/storage"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run consistency checks on the story world",
	Long: `Verify the internal consistency of the story world across three dimensions:

  Character Drift:
    Detects unmotivated belief reversals, personalities drifting without
    cause, and conflicts between active and abandoned beliefs.

  Timeline:
    Checks event ordering by chapter, detects temporal contradictions,
    and flags events where the protagonist is absent without explanation.

  Forgotten Threads:
    Identifies unresolved plot hooks that have been stalled, hooks with
    "immediate" urgency that remain unaddressed, and duplicated hook IDs.

Issues are reported with severity levels:
  ❌ ERROR   — Must fix (breaks continuity)
  ⚠️  WARNING — Should review (potential problem)
  ℹ️  INFO    — For awareness (consider addressing)`,
	RunE: runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
	projectDir := GetProjectDir()
	paths := storage.NewPaths(projectDir)

	if _, err := os.Stat(paths.ProjectYAML()); os.IsNotExist(err) {
		return fmt.Errorf("no project found at %s — run 'ff init' first", projectDir)
	}

	loader := storage.NewLoader(paths)

	protagonist, err := loader.LoadProtagonist()
	if err != nil {
		return fmt.Errorf("load protagonist: %w", err)
	}

	world, err := loader.LoadWorld()
	if err != nil {
		return fmt.Errorf("load world: %w", err)
	}

	events, err := loader.LoadTimeline()
	if err != nil {
		return fmt.Errorf("load timeline: %w", err)
	}

	report := check.RunAll(protagonist, world, events)
	report.Print()

	return nil
}
