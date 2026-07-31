package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	projectDir string
	verbose    bool
)

// RootCmd is the base command for ff.
var RootCmd = &cobra.Command{
	Use:   "ff",
	Short: "Fiction Factory - AI-driven novel production engine",
	Long: `Fiction Factory (ff) is an event-driven, agent-based AI novel production engine.

It simulates a novel creation process through two AI agents:
  - Director Agent: plans events, controls pacing, manages world changes
  - Protagonist Agent: embodies the main character's psychology and decisions

Rather than generating text directly from a prompt, Fiction Factory maintains
a consistent, evolving story world where events drive character development
and plot progression.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if verbose {
			fmt.Fprintf(os.Stderr, "[ff] project directory: %s\n", projectDir)
		}
		return nil
	},
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&projectDir, "project", "p", ".", "Project directory path")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(showCmd)
	RootCmd.AddCommand(runCmd)
	RootCmd.AddCommand(checkCmd)
}

// GetProjectDir returns the project directory from the persistent flag.
func GetProjectDir() string {
	return projectDir
}

// IsVerbose returns whether verbose output is enabled.
func IsVerbose() bool {
	return verbose
}
