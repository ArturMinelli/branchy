package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"branchy/internal/mr"
	"branchy/internal/project"
	"branchy/internal/tui"
)

var mrCmd = &cobra.Command{
	Use:   "mr",
	Short: "Create a GitLab MR between two branches",
	RunE: func(cmd *cobra.Command, args []string) error {
		source, _ := cmd.Flags().GetString("source")
		target, _ := cmd.Flags().GetString("target")
		title, _ := cmd.Flags().GetString("title")
		yes, _ := cmd.Flags().GetBool("yes")

		p, err := project.ResolveFromCWD()
		if err != nil {
			return err
		}

		if UseTUI(cmd) {
			return tui.RunMR(p, tui.MROptions{})
		}

		if target != "" && source == "" {
			return fmt.Errorf("--source is required when --target is set")
		}
		if source == "" || target == "" {
			return fmt.Errorf("mr requires --source and --target in non-interactive mode")
		}
		return runMRFlags(p, source, target, title, yes)
	},
}

func runMRFlags(p *project.Project, source, target, title string, yes bool) error {
	if !yes {
		fmt.Printf("Create MR %s → %s? [y/N] ", source, target)
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "y" && line != "yes" {
			return fmt.Errorf("cancelled")
		}
	}

	result, err := mr.Create(p, mr.CreateRequest{
		Source: source,
		Target: target,
		Title:  title,
	})
	if err != nil {
		return err
	}

	switch result.Action {
	case mr.ActionCreated:
		fmt.Printf("Created: %s → %s\n  %s\n", result.Source, result.Target, result.URL)
	case mr.ActionSkipped:
		fmt.Printf("Skipped: %s → %s (%s)\n  %s\n", result.Source, result.Target, result.Message, result.URL)
	case mr.ActionFailed:
		return fmt.Errorf("%s → %s — %s", result.Source, result.Target, result.Message)
	}
	return nil
}

func init() {
	mrCmd.Flags().String("source", "", "Source branch name")
	mrCmd.Flags().String("target", "", "Target branch name")
	mrCmd.Flags().String("title", "", "MR title (auto-generated if omitted)")
	mrCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}
