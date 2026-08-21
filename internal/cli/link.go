package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"branchy/internal/project"
	"branchy/internal/tui"
)

var linkCmd = &cobra.Command{
	Use:   "link [<parent> <child>]",
	Short: "Add a parent→child edge to the branch tree",
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
				return tui.RunLink(p)
			}
			return fmt.Errorf("link requires <parent> <child> in non-interactive mode")
		}

		p, err := project.ResolveFromCWD()
		if err != nil {
			return err
		}
		parent, child := args[0], args[1]
		if err := p.Link(parent, child); err != nil {
			return err
		}
		fmt.Printf("Linked %s → %s\n", parent, child)
		return nil
	},
}
