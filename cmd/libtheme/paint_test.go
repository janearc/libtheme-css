package main

import (
	"testing"

	"github.com/janearc/libtheme-css/spaces/srgb"
)

// TestPaintIsSilentWhenNotATerminal draws nothing when the output is not
// a terminal, as it is not under go test, and nothing under NO_COLOR, so
// no escape and no run of number signs reaches a file or a screen reader.
func TestPaintIsSilentWhenNotATerminal(t *testing.T) {
	pink := srgb.MustHex("#ff6ec7").Swatch()
	if got := paint(pink, 12); got != "" {
		t.Errorf("piped, paint drew %q", got)
	}
	t.Setenv("NO_COLOR", "1")
	if got := paint(pink, 12); got != "" {
		t.Errorf("with NO_COLOR, paint drew %q", got)
	}
}
