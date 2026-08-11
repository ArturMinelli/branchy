package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestAnyFlagChanged(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("from", "", "")
	cmd.Flags().Bool("yes", false, "")

	if anyFlagChanged(cmd) {
		t.Fatal("expected no flags changed initially")
	}

	if err := cmd.Flags().Set("from", "main"); err != nil {
		t.Fatal(err)
	}
	if !anyFlagChanged(cmd) {
		t.Fatal("expected from flag to be changed")
	}
}

func TestUseTUIRequiresTTY(t *testing.T) {
	cmd := &cobra.Command{Use: "sync"}
	cmd.Flags().String("from", "", "")

	// In test environment stdout may or may not be TTY; test flag path deterministically.
	if err := cmd.Flags().Set("from", "develop"); err != nil {
		t.Fatal(err)
	}
	if UseTUI(cmd) {
		t.Fatal("expected UseTUI false when flag is set")
	}
}

func TestUseTUINoFlagsWhenTTY(t *testing.T) {
	if !IsTTY() {
		t.Skip("stdout is not a TTY in this test environment")
	}
	cmd := &cobra.Command{Use: "sync"}
	cmd.Flags().String("from", "", "")
	if !UseTUI(cmd) {
		t.Fatal("expected UseTUI true with no flags in TTY")
	}
}

func TestSyncCmdFlagForcesCLI(t *testing.T) {
	var buf bytes.Buffer
	syncCmd.SetOut(&buf)
	syncCmd.SetErr(&buf)

	if err := syncCmd.Flags().Set("from", "main"); err != nil {
		t.Fatal(err)
	}
	if UseTUI(syncCmd) {
		t.Fatal("sync with --from should not use TUI")
	}

	_ = syncCmd.Flags().Set("from", "")
}
