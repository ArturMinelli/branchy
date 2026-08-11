package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type fetchFlight struct {
	done chan struct{}
	err  error
}

var (
	flightsMu sync.Mutex
	flights   = map[string]*fetchFlight{}
)

// DefaultRemote returns the remote used for background updates: origin if
// configured, otherwise the first remote. ok is false when none exist.
func DefaultRemote(dir string) (string, bool) {
	if dir == "" {
		return "", false
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", false
	}
	out, err := exec.Command("git", "-C", abs, "remote").Output()
	if err != nil {
		return "", false
	}
	var names []string
	for _, line := range strings.Split(string(out), "\n") {
		name := strings.TrimSpace(line)
		if name != "" {
			names = append(names, name)
		}
	}
	for _, name := range names {
		if name == "origin" {
			return "origin", true
		}
	}
	if len(names) > 0 {
		return names[0], true
	}
	return "", false
}

// FetchDefaultRemote updates tracking refs for the repo's default remote.
// Concurrent callers for the same directory share one in-flight fetch.
// No remotes is a successful no-op.
func FetchDefaultRemote(dir string) error {
	if dir == "" {
		return fmt.Errorf("dir is required")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	flightsMu.Lock()
	if f, ok := flights[abs]; ok {
		flightsMu.Unlock()
		<-f.done
		return f.err
	}
	f := &fetchFlight{done: make(chan struct{})}
	flights[abs] = f
	flightsMu.Unlock()

	err = doFetch(abs)
	f.err = err
	close(f.done)

	flightsMu.Lock()
	delete(flights, abs)
	flightsMu.Unlock()
	return err
}

func doFetch(dir string) error {
	remote, ok := DefaultRemote(dir)
	if !ok {
		return nil
	}
	out, err := exec.Command("git", "-C", dir, "fetch", remote).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch %s: %w: %s", remote, err, strings.TrimSpace(string(out)))
	}
	return nil
}
