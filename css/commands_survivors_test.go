package css_test

import (
	"slices"
	"testing"

	"github.com/janearc/libtheme-css/css"
)

// These tests hold what the mutation run found nothing holding in the
// reader: that a comment blanked out keeps its line ends, so a line number
// named in a complaint is the file's.

// Read names the file's line for a bad colour after a long comment. The
// anti-test is the same colour before it.
func TestSkippedNamesTheLineAfterAMultiLineComment(t *testing.T) {
	after := css.Read("/* one\ntwo\nthree */\n:root { --bad: #12; }\n")
	wantAfter := []string{"line 4: --bad: #12 is not a #rgb or #rrggbb"}
	if !slices.Equal(after.Skipped(), wantAfter) {
		t.Errorf("after a comment: %v", after.Skipped())
	}
	before := css.Read(":root { --bad: #12; }\n/* one\ntwo\n*/\n")
	wantBefore := []string{"line 1: --bad: #12 is not a #rgb or #rrggbb"}
	if !slices.Equal(before.Skipped(), wantBefore) {
		t.Errorf("before a comment: %v", before.Skipped())
	}
}
