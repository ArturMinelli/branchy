package browser

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Open tries to open a URL in the system browser.
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
	return cmd.Start()
}
