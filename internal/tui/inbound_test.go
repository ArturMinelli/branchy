package tui

import "testing"

func TestFormatInboundBadge(t *testing.T) {
	if got := formatInboundBadge(inboundCount{files: 12, ok: true}); got != "12" {
		t.Fatalf("positive: got %q", got)
	}
	if got := formatInboundBadge(inboundCount{files: 0, ok: true}); got != "" {
		t.Fatalf("zero must be hidden, got %q", got)
	}
	if got := formatInboundBadge(inboundCount{ok: false}); got != "?" {
		t.Fatalf("unknown: got %q", got)
	}
}

func TestFormatInboundConfirm(t *testing.T) {
	if got := formatInboundConfirm("feature", inboundCount{files: 4, ok: true}); got != "4 files would change on feature" {
		t.Fatalf("positive: got %q", got)
	}
	if got := formatInboundConfirm("feature", inboundCount{files: 0, ok: true}); got != "0 files would change on feature" {
		t.Fatalf("known zero: got %q", got)
	}
	if got := formatInboundConfirm("feature", inboundCount{ok: false}); got != "File count unavailable" {
		t.Fatalf("unknown: got %q", got)
	}
}
