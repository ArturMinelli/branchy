package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"branchy/internal/project"
	"branchy/internal/tui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Register the current git repo with branchy",
	RunE: func(cmd *cobra.Command, args []string) error {
		if UseTUI(cmd) {
			return tui.RunInit()
		}
		force, _ := cmd.Flags().GetBool("force")
		p, err := project.Init(project.InitOptions{Force: force})
		if err != nil {
			return err
		}
		fmt.Printf("Registered %q at %s\n", p.ID, p.Path)
		return nil
	},
}

func init() {
	initCmd.Flags().Bool("force", false, "Re-import branch tree for an already registered repo")
}
