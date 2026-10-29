package vim_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/dialects/vim"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// eight is the colours in each row of vim's terminal table. It is
// defined by the sixteen: eight and their bright eight.
const eight = 8

// quoted is the colour of dusk as the terminal table writes it.
const quoted = "'" + grey + "'"

// Write puts the sixteen for vim's terminal in two rows, split at the
// eighth colour.
func TestTerminalIsTwoRowsOfEight(t *testing.T) {
	text, err := written(t, "dusk", "h")
	if err != nil {
		t.Fatal(err)
	}
	items := slices.Repeat([]string{quoted}, eight)
	row := "  \\ " + strings.Join(items, ", ")
	want := "let g:terminal_ansi_colors = [\n" + row + ",\n" + row + "]\n"
	if !strings.Contains(text, want) {
		t.Errorf("the terminal table is not two rows of eight in\n%s", text)
	}
}

// Write puts a group's colour only when the group names it and the
// scheme holds it. No name gets none. This holds even when the scheme
// has a colour under the empty name.
func TestAGroupWithNoColourNamedGetsNone(t *testing.T) {
	scheme := dusk(t, "dusk")
	scheme.Colours[""] = srgb.MustHex(grey).Swatch()
	var out strings.Builder
	if err := vim.Write(&out, scheme, "h"); err != nil {
		t.Fatal(err)
	}
	for line := range strings.SplitSeq(out.String(), "\n") {
		switch {
		case strings.HasPrefix(line, "hi CursorLine "):
			if strings.Contains(line, "guifg") {
				t.Errorf("a group with no text got one: %q", line)
			}
		case strings.HasPrefix(line, "hi LineNr "):
			if strings.Contains(line, "guibg") {
				t.Errorf("a group with no ground got one: %q", line)
			}
		case strings.HasPrefix(line, "hi Normal "):
			if !strings.Contains(line, "guifg") || !strings.Contains(line, "guibg") {
				t.Errorf("a group with both lost one: %q", line)
			}
		}
	}
}
