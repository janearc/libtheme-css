package nvim_test

import (
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/dialects/nvim"
)

// nothing is no bytes, or no colours. It is defined by counting.
const nothing = 0

// partial is a scheme file with a colour table and no terminal table, which
// Read accepts.
const partial = "vim.g.colors_name = \"half\"\nlocal c = {\n  bg = \"#101010\",\n  ink = \"#c0c0c0\",\n}\n"

// The anti-tests: a scheme read from a fragment is not ready to write, and
// writing it is an error before any byte is written, not a panic.
func TestAPartialSchemeIsNotWritten(t *testing.T) {
	scheme, err := nvim.Read(strings.NewReader(partial))
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if err := scheme.Write(&out, "header"); err == nil || out.Len() != nothing {
		t.Errorf("a partial scheme was written (%v), %d bytes", err, out.Len())
	}
	var empty nvim.Scheme
	empty.Name = "empty"
	if err := empty.Write(&out, "header"); err == nil || out.Len() != nothing {
		t.Errorf("a scheme of nothing was written (%v), %d bytes", err, out.Len())
	}
}

// A scheme with all its colours and a short terminal table is refused too,
// and so is one with a colour missing from a table of sixteen.
func TestAShortOrHolyScheme(t *testing.T) {
	whole := dusk(t, "dusk")
	short := dusk(t, "dusk")
	short.Terminal = short.Terminal[:len(short.Terminal)/half]
	holy := dusk(t, "dusk")
	delete(holy.Colours, "panel")
	var out strings.Builder
	for name, scheme := range map[string]nvim.Scheme{"short": short, "holy": holy} {
		if err := scheme.Write(&out, "header"); err == nil {
			t.Errorf("a %s scheme was written", name)
		}
	}
	if err := whole.Write(&out, "header"); err != nil {
		t.Errorf("a whole scheme was refused: %v", err)
	}
}

// Complete says which colour is missing, by the name the file uses.
func TestCompleteNamesTheMissingColour(t *testing.T) {
	holy := dusk(t, "dusk")
	delete(holy.Colours, "panel")
	err := holy.Complete()
	if err == nil || !strings.Contains(err.Error(), "panel") {
		t.Errorf("the missing colour was not named: %v", err)
	}
	short := dusk(t, "dusk")
	short.Terminal = nil
	if err := short.Complete(); err == nil || !strings.Contains(err.Error(), "sixteen") {
		t.Errorf("the missing sixteen was not named: %v", err)
	}
}

// The property: whichever of the colours and however many of the sixteen a
// scheme has, writing it never panics, and it is written exactly when it
// has them all.
func TestWritingNeverPanicsAndIsWholeOrNothing(t *testing.T) {
	rapid.Check(t, writesWholeOrNothing)
}

// writesWholeOrNothing is one draw of the property above.
func writesWholeOrNothing(t *rapid.T) {
	scheme := dusk(t, "dusk")
	for _, entry := range nvim.Roles {
		if !rapid.Bool().Draw(t, "has "+entry.Name) {
			delete(scheme.Colours, entry.Name)
		}
	}
	kept := rapid.IntRange(nothing, len(scheme.Terminal)).Draw(t, "sixteen")
	scheme.Terminal = scheme.Terminal[:kept]
	var out strings.Builder
	err := scheme.Write(&out, "header")
	complete := len(scheme.Colours) == len(nvim.Roles) && len(scheme.Terminal) == wholeTerminal
	if (err == nil) != complete {
		t.Fatalf("error %v for %d colours and %d of the sixteen", err,
			len(scheme.Colours), len(scheme.Terminal))
	}
	if err != nil && out.Len() != nothing {
		t.Fatalf("%d bytes were written before the refusal", out.Len())
	}
}

// wholeTerminal is how many colours the terminal's table has. It is defined
// by the ansi palette.
const wholeTerminal = 16

// half cuts a table in two. It is defined by halving.
const half = 2
