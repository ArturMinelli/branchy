package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"branchy/internal/project"
	"branchy/internal/sync"
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
	rootCmd.AddCommand(linkCmd)
	rootCmd.AddCommand(projectsCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Register the current git repo with branchy",
	RunE: func(cmd *cobra.Command, args []string) error {
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

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Create GitLab MRs along branch tree edges",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		yes, _ := cmd.Flags().GetBool("yes")

		p, err := project.ResolveFromCWD()
		if err != nil {
			return err
		}

		if from == "" {
			names := p.Tree.Names()
			if len(names) == 0 {
				return fmt.Errorf("branch tree is empty")
			}
			fmt.Println("Select root branch to sync from:")
			for i, name := range names {
				fmt.Printf("  %d) %s\n", i+1, name)
			}
			fmt.Print("> ")
			reader := bufio.NewReader(os.Stdin)
			line, _ := reader.ReadString('\n')
			line = strings.TrimSpace(line)
			var idx int
			if _, err := fmt.Sscanf(line, "%d", &idx); err != nil || idx < 1 || idx > len(names) {
				return fmt.Errorf("invalid selection")
			}
			from = names[idx-1]
		}

		edges := p.Tree.CollectEdges(from)
		if len(edges) == 0 {
			fmt.Printf("No child branches below %q.\n", from)
			return nil
		}

		fmt.Println("Sync plan (parent → child):")
		for _, e := range edges {
			fmt.Printf("  • %s → %s\n", e.Parent, e.Child)
		}
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		confirm := func(parent, child string) (bool, error) {
			if yes {
				return true, nil
			}
			fmt.Printf("Create MR %s → %s? [y/N] ", parent, child)
			line, err := reader.ReadString('\n')
			if err != nil {
				return false, err
			}
			line = strings.TrimSpace(strings.ToLower(line))
			return line == "y" || line == "yes", nil
		}

		summary, err := sync.Run(p, sync.Options{
			FromBranch: from,
			Confirm:    confirm,
			OnStatus: func(msg string) { fmt.Println(msg) },
		})
		if err != nil {
			return err
		}

		created, skipped, failed := 0, 0, 0
		for _, r := range summary.Results {
			switch r.Action {
			case "created":
				created++
				fmt.Printf("Created: %s → %s\n  %s\n", r.Parent, r.Child, r.URL)
			case "skipped":
				skipped++
				fmt.Printf("Skipped: %s → %s (%s)\n", r.Parent, r.Child, r.Message)
			case "failed":
				failed++
				fmt.Printf("Failed: %s → %s — %s\n", r.Parent, r.Child, r.Message)
			}
		}
		fmt.Printf("\nDone — created: %d, skipped: %d, failed: %d\n", created, skipped, failed)
		return nil
	},
}

func init() {
	syncCmd.Flags().String("from", "", "Root branch to sync from")
	syncCmd.Flags().BoolP("yes", "y", false, "Create all MRs without prompting")
}

var linkCmd = &cobra.Command{
	Use:   "link <parent> <child>",
	Short: "Add a parent→child edge to the branch tree",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := project.ResolveFromCWD()
		if err != nil {
			return err
		}
		parent, child := args[0], args[1]
		if err := p.Tree.Link(parent, child); err != nil {
			return err
		}
		if err := p.SaveTree(); err != nil {
			return err
		}
		fmt.Printf("Linked %s → %s\n", parent, child)
		return nil
	},
}

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List registered projects",
	RunE: func(cmd *cobra.Command, args []string) error {
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
