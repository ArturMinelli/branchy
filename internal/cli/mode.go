package cli

import (
	"os"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// IsTTY reports whether stdout is an interactive terminal.
func IsTTY() bool {
	return term.IsTerminal(os.Stdout.Fd())
}

// UseTUI reports whether a command should launch the interactive TUI.
// Returns true only when stdout is a TTY and no flags were explicitly set.
func UseTUI(cmd *cobra.Command) bool {
	return IsTTY() && !anyFlagChanged(cmd)
}

func anyFlagChanged(cmd *cobra.Command) bool {
	changed := false
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if f.Changed {
			changed = true
		}
	})
	return changed
}
