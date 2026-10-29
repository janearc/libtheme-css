package ghostty_test

import (
	"fmt"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/dialects/ghostty"
)

// The numbers the palette tests pick.
const (
	// lowestEntry is the first palette entry. It is defined by ghostty: the
	// palette is a byte's worth of entries.
	lowestEntry = 0
	// highestEntry is the last palette entry. It is defined by ghostty: the
	// palette is a byte's worth of entries.
	highestEntry = 255
	// reach is how far past the palette a drawn entry may fall. It is
	// defined by this test.
	reach = 1000
)

// anyColour is a palette colour that parses.
const anyColour = "#102030"

// How many entries a theme has, counted.
const (
	// nothing is no entries. It is defined by counting.
	nothing = 0
	// one is a single entry. It is defined by counting.
	one = 1
)

// read is a theme from the text of its file.
func read(text string) (ghostty.Theme, error) {
	return ghostty.Read(strings.NewReader(text))
}

// The anti-test: a palette line with something wrong in its entry is an
// error that names the line, and the theme it would have made is not given.
func TestReadRefusesABadPaletteEntry(t *testing.T) {
	for name, line := range map[string]string{
		"no equals sign":   "palette = 5",
		"not a number":     "palette = five=" + anyColour,
		"below the first":  "palette = -1=" + anyColour,
		"past the last":    "palette = 256=" + anyColour,
		"nothing for it":   "palette = =" + anyColour,
		"a fraction":       "palette = 1.5=" + anyColour,
		"two entries in 1": "palette = 1=2=" + anyColour,
	} {
		theme, err := read("# theme\n" + line + "\n")
		if err == nil || !strings.Contains(err.Error(), "line 2") {
			t.Errorf("%s: got error %v", name, err)
		}
		if len(theme.Palette) != nothing {
			t.Errorf("%s: a theme came back with %d entries", name,
				len(theme.Palette))
		}
	}
}

// The anti-test: a palette entry whose colour does not parse is an error that
// names the line.
func TestReadRefusesABadPaletteColour(t *testing.T) {
	_, err := read(
		"background = " + anyColour + "\npalette = 3=brown\n")
	if err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Errorf("got %v", err)
	}
}

// The property: a palette entry is taken when it is a byte's worth, from the
// first to the last, and refused anywhere else, and a taken entry is the one
// named.
func TestReadTakesExactlyTheBytesWorthOfEntries(t *testing.T) {
	rapid.Check(t, takesTheBytesWorth)
}

// takesTheBytesWorth is one draw of the property above.
func takesTheBytesWorth(t *rapid.T) {
	entry := rapid.IntRange(lowestEntry-reach, highestEntry+reach).Draw(t,
		"entry")
	theme, err := read(
		fmt.Sprintf("palette = %d=%s\n", entry, anyColour))
	inside := entry >= lowestEntry && entry <= highestEntry
	if (err == nil) != inside {
		t.Fatalf("entry %d: error %v, but the palette %s it", entry, err,
			map[bool]string{true: "has", false: "lacks"}[inside])
	}
	if inside {
		if _, found := theme.Palette[entry]; !found || len(theme.Palette) != one {
			t.Fatalf("entry %d: palette %v", entry, theme.Palette)
		}
	}
}
