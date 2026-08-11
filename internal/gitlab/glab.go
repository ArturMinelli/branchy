package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Client shells out to glab in a project directory.
type Client struct {
	Dir string
}

// AuthOK returns nil when glab is authenticated for the project.
func (c *Client) AuthOK() error {
	_, err := c.run("auth", "status")
	return err
}

// FindOpenMR returns the web URL of an open MR for source→target, or "".
func (c *Client) FindOpenMR(source, target string) (string, error) {
	out, err := c.run(
		"mr", "list",
		"--source-branch", source,
		"--target-branch", target,
		"--state", "opened",
		"--per-page", "1",
		"-F", "json",
	)
	if err != nil {
		return "", err
	}
	out = strings.TrimSpace(out)
	if out == "" || out == "[]" || out == "null" {
		return "", nil
	}

	var items []map[string]any
	if err := json.Unmarshal([]byte(out), &items); err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", nil
	}
	if url, ok := items[0]["web_url"].(string); ok {
		return url, nil
	}
	return "", nil
}

// CreateMR creates a merge request and returns its web URL.
// If creation fails because an open MR already exists, it recovers that URL.
func (c *Client) CreateMR(source, target, title, description string) (string, error) {
	out, err := c.run(
		"mr", "create",
		"--source-branch", source,
		"--target-branch", target,
		"--title", title,
		"--description", description,
		"--yes",
	)
	if err != nil {
		if existing, findErr := c.FindOpenMR(source, target); findErr == nil && existing != "" {
			return existing, errAlreadyExists
		}
		return "", err
	}

	url := extractURL(out)
	if url != "" {
		return url, nil
	}

	return c.FindOpenMR(source, target)
}

// errAlreadyExists is returned by CreateMR when an open MR was recovered after a create failure.
var errAlreadyExists = fmt.Errorf("open MR already exists")

// IsAlreadyExists reports whether err indicates an existing open MR was recovered.
func IsAlreadyExists(err error) bool {
	return err != nil && errors.Is(err, errAlreadyExists)
}

func (c *Client) run(args ...string) (string, error) {
	cmd := exec.Command("glab", args...)
	cmd.Dir = c.Dir
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		if text != "" {
			return "", fmt.Errorf("%w: %s", err, text)
		}
		return "", err
	}
	return text, nil
}

var urlRE = regexp.MustCompile(`https?://[^\s]+`)

func extractURL(text string) string {
	matches := urlRE.FindAllString(text, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1]
}
