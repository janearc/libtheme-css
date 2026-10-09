package ghostty

import (
	"fmt"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// colourway is a small complete one: the ground and ink, the selection,
// the sixteen, and two numbered entries above them.
func colourway() *css.Sheet {
	roles := css.New()
	roles.Set("ground", srgb.MustHex("#808a63").Swatch())
	roles.Set("ink", srgb.MustHex("#38241e").Swatch())
	roles.Set("surface-1", srgb.MustHex("#7f794b").Swatch())
	for number, name := range Ansi {
		roles.Set(name, srgb.RGB8(uint8(number*16), 40, uint8(255-number*16)).Swatch())
	}
	roles.Set("palette-252", srgb.MustHex("#c5e699").Swatch())
	roles.Set("palette-188", srgb.MustHex("#20e1e7").Swatch())
	return roles
}

// written is a theme as its file.
func written(t *testing.T, theme Theme) string {
	t.Helper()
	var out strings.Builder
	if err := theme.Write(&out, "a test theme\nin two lines"); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

// a colourway makes the file it says it does: each setting by its role,
// the sixteen in order, the numbered entries after them.
func TestOfWritesTheRoles(t *testing.T) {
	theme, err := Of(colourway())
	if err != nil {
		t.Fatal(err)
	}
	text := written(t, theme)
	for _, line := range []string{"# a test theme", "# in two lines",
		"background = #808a63", "foreground = #38241e",
		"selection-background = #7f794b", "palette = 0=#0028ff",
		"palette = 188=#20e1e7", "palette = 252=#c5e699"} {
		if !strings.Contains(text, line+"\n") {
			t.Errorf("no %q in\n%s", line, text)
		}
	}
	if strings.Index(text, "palette = 188") > strings.Index(text, "palette = 252") {
		t.Error("the numbered entries are not in order")
	}
}

// a file this dialect wrote reads back to the same file, byte for byte:
// nothing is lost or moved on the way through the swatches.
func TestReadWriteIsIdentity(t *testing.T) {
	theme, err := Of(colourway())
	if err != nil {
		t.Fatal(err)
	}
	first := written(t, theme)
	again, err := Read(strings.NewReader(first))
	if err != nil {
		t.Fatal(err)
	}
	if second := written(t, again); second != first {
		t.Errorf("read and written again, the file changed:\n%s\n---\n%s", first, second)
	}
}

// the roles a theme says back make the same theme.
func TestRolesRoundTrip(t *testing.T) {
	theme, err := Of(colourway())
	if err != nil {
		t.Fatal(err)
	}
	back, err := Of(theme.Roles())
	if err != nil {
		t.Fatal(err)
	}
	if written(t, back) != written(t, theme) {
		t.Error("the theme's own roles make a different theme")
	}
}

// a colourway without its ground, its ink or any of the sixteen is
// refused by name, and a number below 16 is sent to its name.
func TestOfRefusesWhatIsMissing(t *testing.T) {
	for _, missing := range []string{"ground", "ink", "bright-cyan"} {
		roles := css.New()
		for _, rule := range colourway().Rules() {
			if rule.Name != missing {
				roles.Set(rule.Name, rule.Swatch)
			}
		}
		if _, err := Of(roles); err == nil || !strings.Contains(err.Error(), missing) {
			t.Errorf("without %s: %v", missing, err)
		}
	}
	roles := colourway()
	roles.Set("palette-3", srgb.MustHex("#ffffff").Swatch())
	if _, err := Of(roles); err == nil {
		t.Error("palette-3 was taken as a number")
	}
}

// a colour that does not parse is an error that names its line.
func TestReadNamesTheBadLine(t *testing.T) {
	_, err := Read(strings.NewReader("# a theme\nbackground = #808a63\nforeground = brown\n"))
	if err == nil || !strings.Contains(err.Error(), "line 3") {
		t.Errorf("got %v", err)
	}
}

// inverted twice is the theme again, and once is each lamp's opposite.
func TestInvertedTwiceIsTheTheme(t *testing.T) {
	theme, err := Of(colourway())
	if err != nil {
		t.Fatal(err)
	}
	if written(t, theme.Inverted().Inverted()) != written(t, theme) {
		t.Error("inverting twice changed the theme")
	}
	if got := Hex(theme.Inverted().Settings["background"]); got != "#7f759c" {
		t.Errorf("the inverted ground is %s, not #7f759c", got)
	}
}

// every byte survives the trip from hex into a swatch and back.
func TestEveryByteSurvives(t *testing.T) {
	for level := range 256 {
		hex := fmt.Sprintf("#%02x%02x%02x", level, 255-level, level/2)
		value, err := parse(hex)
		if err != nil {
			t.Fatal(err)
		}
		if Hex(value) != hex {
			t.Errorf("%s came back as %s", hex, Hex(value))
		}
	}
}

// a partial file is written back partial: nothing is invented for the
// entries it does not have, and ghostty's own words for a colour are
// passed over.
func TestPartialStaysPartial(t *testing.T) {
	partial := "# part of a theme\n\nbackground = #808a63\n" +
		"cursor-color = cell-foreground\n\npalette = 0=#57696a\n"
	theme, err := Read(strings.NewReader(partial))
	if err != nil {
		t.Fatal(err)
	}
	text := written(t, theme)
	if strings.Contains(text, "palette = 1=") {
		t.Errorf("an entry was invented:\n%s", text)
	}
	if strings.Contains(text, "cursor-color") {
		t.Errorf("a cell colour became a colour:\n%s", text)
	}
}
