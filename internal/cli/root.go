package cli

import (
	"github.com/spf13/cobra"

	"branchy/internal/project"
	"branchy/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "branchy",
	Short: "Branch hierarchy TUI and GitLab MR sync",
	Long:  "Visualize branch trees and create GitLab merge requests along parent→child edges.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var preselected *project.Project
		if p, err := project.ResolveFromCWD(); err == nil {
			preselected = p
		}
		return tui.Run(preselected)
	},
}

// Execute runs the CLI.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(mrCmd)
	rootCmd.AddCommand(linkCmd)
	rootCmd.AddCommand(unlinkCmd)
	rootCmd.AddCommand(projectsCmd)
}
