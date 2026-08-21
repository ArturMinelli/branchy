package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"branchy/internal/project"
	"branchy/internal/tui"
)

var unlinkCmd = &cobra.Command{
	Use:   "unlink [<parent> <child>]",
	Short: "Remove a child subtree from the branch tree",
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return fmt.Errorf("requires 0 or 2 arguments")
		}
		if len(args) == 0 {
			if UseTUI(cmd) {
				p, err := project.ResolveFromCWD()
				if err != nil {
					return err
				}
				return tui.RunUnlink(p)
			}
			return fmt.Errorf("unlink requires <parent> <child> in non-interactive mode")
		}

		p, err := project.ResolveFromCWD()
		if err != nil {
			return err
		}
		parent, child := args[0], args[1]
		res, err := p.Unlink(parent, child)
		if err != nil {
			return err
		}
		fmt.Printf("Unlinked %s → %s (%d branches removed)\n", res.Parent, res.Child, res.Removed)
		return nil
	},
}
