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

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Create GitLab MRs along branch tree edges",
	RunE: func(cmd *cobra.Command, args []string) error {
		if UseTUI(cmd) {
			p, err := project.ResolveFromCWD()
			if err != nil {
				return err
			}
			return tui.RunSync(p, tui.SyncFlowOptions{})
		}
		return runSyncCLI(cmd)
	},
}

func runSyncCLI(cmd *cobra.Command) error {
	from, _ := cmd.Flags().GetString("from")
	yes, _ := cmd.Flags().GetBool("yes")

	p, err := project.ResolveFromCWD()
	if err != nil {
		return err
	}

	if from == "" {
		if !IsTTY() {
			return fmt.Errorf("sync requires --from in non-interactive mode")
		}
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
			if r.URL != "" {
				fmt.Printf("  %s\n", r.URL)
			}
		case "failed":
			failed++
			fmt.Printf("Failed: %s → %s — %s\n", r.Parent, r.Child, r.Message)
		}
	}
	fmt.Printf("\nDone — created: %d, skipped: %d, failed: %d\n", created, skipped, failed)

	if urls := sync.OpenableURLs(summary); len(urls) > 0 {
		fmt.Print("Open MRs in browser? [y/N] ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(strings.ToLower(line))
		if line == "y" || line == "yes" {
			fmt.Printf("Opening %d MR(s) in browser...\n", len(urls))
			if warn := sync.OpenURLs(urls); warn != "" {
				fmt.Printf("Warning: %s\n", warn)
			}
		}
	}

	return nil
}

func init() {
	syncCmd.Flags().String("from", "", "Root branch to sync from")
	syncCmd.Flags().BoolP("yes", "y", false, "Create all MRs without prompting")
}
