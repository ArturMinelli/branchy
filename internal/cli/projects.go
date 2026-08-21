package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"branchy/internal/project"
	"branchy/internal/tui"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List registered projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		if UseTUI(cmd) {
			return tui.RunProjects()
		}
		projects, err := project.ListAll()
		if err != nil {
			return err
		}
		if len(projects) == 0 {
			fmt.Println("No projects registered.")
			return nil
		}
		for _, p := range projects {
			fmt.Printf("%s\t%s\n", p.ID, p.Path)
		}
		return nil
	},
}
