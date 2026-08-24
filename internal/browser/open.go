package browser

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var openURL = Open

// Open tries to open a URL in the system browser.
// The child process is detached from the caller's process group so that
// exiting a TUI (alt-screen) does not kill xdg-open before tabs open.
func Open(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("xdg-open"); err == nil {
			cmd = exec.Command("xdg-open", url)
		}
	case "darwin":
		cmd = exec.Command("open", url)
	}

	if cmd == nil {
		if browser := os.Getenv("BROWSER"); browser != "" {
			cmd = exec.Command(browser, url)
		}
	}

	if cmd == nil {
		return fmt.Errorf("could not open browser; URL: %s", url)
	}

	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	}
	return cmd.Start()
}

// OpenURLs opens each URL sequentially in order. Returns a non-fatal warning if any open fails.
func OpenURLs(urls []string) string {
	if len(urls) == 0 {
		return ""
	}
	var warnings []string
	for i, url := range urls {
		if i > 0 {
			time.Sleep(200 * time.Millisecond)
		}
		if err := openURL(url); err != nil {
			warnings = append(warnings, fmt.Sprintf("could not open %s: %v", url, err))
		}
	}
	return strings.Join(warnings, "; ")
}
