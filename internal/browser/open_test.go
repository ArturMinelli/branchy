package browser

import (
	"errors"
	"testing"
)

func TestOpenURLsEmpty(t *testing.T) {
	if warn := OpenURLs(nil); warn != "" {
		t.Fatalf("expected empty warning, got %q", warn)
	}
}

func TestOpenURLsSequentialAndWarning(t *testing.T) {
	var opened []string
	orig := openURL
	t.Cleanup(func() { openURL = orig })
	openURL = func(url string) error {
		opened = append(opened, url)
		if url == "bad" {
			return errors.New("nope")
		}
		return nil
	}

	warn := OpenURLs([]string{"first", "bad", "third"})
	if len(opened) != 3 {
		t.Fatalf("expected 3 opens, got %d: %v", len(opened), opened)
	}
	if opened[0] != "first" || opened[1] != "bad" || opened[2] != "third" {
		t.Fatalf("unexpected open order: %v", opened)
	}
	if warn == "" {
		t.Fatal("expected warning for failed open")
	}
}
