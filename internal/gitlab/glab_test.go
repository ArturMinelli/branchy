package gitlab

import "testing"

func TestOpenMRListArgsOmitsStateFlag(t *testing.T) {
	args := openMRListArgs("develop", "feature")

	for _, arg := range args {
		if arg == "--state" {
			t.Fatalf("glab mr list has no --state flag, got %v", args)
		}
	}

	want := []string{
		"mr", "list",
		"--source-branch", "develop",
		"--target-branch", "feature",
		"--per-page", "1",
		"-F", "json",
	}
	if len(args) != len(want) {
		t.Fatalf("got %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("index %d: got %q, want %q", i, args[i], want[i])
		}
	}
}
