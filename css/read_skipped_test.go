package css_test

import (
	"slices"
	"testing"

	"github.com/janearc/libtheme-css/css"
)

// These tests hold what a complaint from Skipped says: the line the
// property's name is on, and the value as the file wrote it.

// skippedOf is what Skipped says of a source.
func skippedOf(src string) []string {
	return css.Read(src).Skipped()
}

// The line is where the name starts, not where the value ends. Values on
// the next line do not move it. Nor does a missing semicolon.
func TestSkippedNamesTheLineTheNameIsOn(t *testing.T) {
	for src, want := range map[string]string{
		":root {\n  --x:\n    #12;\n}\n": "line 2: --x: #12 is not a #rgb or #rrggbb",
		":root {\n  --x: #12\n}\n":       "line 2: --x: #12 is not a #rgb or #rrggbb",
		"--x: #12\n}":                    "line 1: --x: #12 is not a #rgb or #rrggbb",
		"\n\n:root { --x: #12; }":        "line 3: --x: #12 is not a #rgb or #rrggbb",
	} {
		if got := skippedOf(src); !slices.Equal(got, []string{want}) {
			t.Errorf("%q: %q, want %q", src, got, want)
		}
	}
}

// The complaint quotes the file's own text for the value, comments and all.
// Search the file for it and the place is found. Comments inside a value
// are quoted whole, close mark too.
func TestSkippedQuotesTheValueAsWritten(t *testing.T) {
	for src, want := range map[string]string{
		":root { --x: #12/**/3; }":      "line 1: --x: #12/**/3 is not a #rgb or #rrggbb",
		":root { --x: #12/* ; */3; }":   "line 1: --x: #12/* ; */3 is not a #rgb or #rrggbb",
		":root { --x: #12 /* never":     "line 1: --x: #12 /* never is not a #rgb or #rrggbb",
		":root { --x: #12 /* it */ ; }": "line 1: --x: #12 /* it */ is not a #rgb or #rrggbb",
	} {
		if got := skippedOf(src); !slices.Equal(got, []string{want}) {
			t.Errorf("%q: %q, want %q", src, got, want)
		}
	}
}

// Comments are blanked a byte at a time. Text before a declaration holds
// its place whatever the comment is made of. The quote is the right bytes.
func TestSkippedQuotesTheRightBytesAfterAWideComment(t *testing.T) {
	src := "/* caf\u00e9, \u2603 */ :root { --x: #12/* \u00e9 */3; }"
	want := "line 1: --x: #12/* \u00e9 */3 is not a #rgb or #rrggbb"
	if got := skippedOf(src); !slices.Equal(got, []string{want}) {
		t.Errorf("%q, want %q", got, want)
	}
}

// Complete colours are not complained of, comment or not. This is the
// anti-test to the quoting above.
func TestACompleteColourBesideACommentIsNotComplainedOf(t *testing.T) {
	if got := skippedOf(":root { --x: #123 /* three */; }"); len(got) != nothing {
		t.Errorf("complained: %q", got)
	}
}

// The flag after a colour may be in any case. It may have space after the
// bang. Colours with it are read, as white. The anti-tests are not flags.
// One has no bang. One has a longer word. One has the flag twice. One has
// a bang and nothing after it. Each is complained of.
func TestTheImportantFlagIsReadInAnyCase(t *testing.T) {
	for _, flag := range []string{"!important", "! important", "!IMPORTANT",
		"!Important", "!  important", " !important"} {
		read := css.Read(":root { --x: #FFF " + flag + "; }")
		if got := read.Skipped(); len(got) != nothing {
			t.Errorf("%q complained: %q", flag, got)
		}
		if names := read.Roles.Names(); !slices.Equal(names, []string{"x"}) {
			t.Errorf("%q: roles %v", flag, names)
		}
	}
	for _, flag := range []string{"important", "!importantly", "!important !important",
		"!"} {
		read := css.Read(":root { --x: #FFF " + flag + "; }")
		if got := read.Skipped(); len(got) == nothing {
			t.Errorf("%q was not complained of", flag)
		}
	}
}
