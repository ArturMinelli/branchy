package sync

import (
	"errors"
	"testing"
)

func TestCreatedURLsFiltersAndPreservesOrder(t *testing.T) {
	summary := &Summary{
		Results: []Result{
			{Parent: "a", Child: "b", Action: "created", URL: "https://example.com/1"},
			{Parent: "b", Child: "c", Action: "skipped", URL: "https://example.com/2"},
			{Parent: "c", Child: "d", Action: "created", URL: "https://example.com/3"},
			{Parent: "d", Child: "e", Action: "failed", URL: ""},
		},
	}

	got := CreatedURLs(summary)
	want := []string{"https://example.com/1", "https://example.com/3"}
	if len(got) != len(want) {
		t.Fatalf("expected %d URLs, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestCreatedURLsNilSummary(t *testing.T) {
	if got := CreatedURLs(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestOpenURLsEmpty(t *testing.T) {
	if warn := OpenURLs(nil); warn != "" {
		t.Fatalf("expected empty warning, got %q", warn)
	}
}

func TestOpenURLsSequentialAndWarning(t *testing.T) {
	var opened []string
	orig := browserOpen
	t.Cleanup(func() { browserOpen = orig })
	browserOpen = func(url string) error {
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
