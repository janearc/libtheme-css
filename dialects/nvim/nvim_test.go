package nvim

import (
	"math"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// colourway is every role a scheme needs, and the sixteen.
func colourway() *css.Sheet {
	roles := css.New()
	for position, entry := range Roles {
		roles.Set(entry.Role, srgb.RGB8(uint8(position*15), 90, uint8(200-position*10)).Swatch())
	}
	roles.Set("ground", srgb.MustHex("#808a63").Swatch())
	for number, name := range ghostty.Ansi {
		roles.Set(name, srgb.RGB8(uint8(number*16), 40, uint8(255-number*16)).Swatch())
	}
	return roles
}

// written is a scheme as its file.
func written(t *testing.T, scheme Scheme) string {
	t.Helper()
	var out strings.Builder
	if err := scheme.Write(&out, "a test scheme"); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

// a colourway makes a whole scheme: its name, its table with the
// contrast beside each colour, the fixed groups, and the sixteen.
func TestOfWritesTheScheme(t *testing.T) {
	scheme, err := Of("test-dusk", colourway())
	if err != nil {
		t.Fatal(err)
	}
	text := written(t, scheme)
	for _, line := range []string{"-- a test scheme", `vim.g.colors_name = "test-dusk"`,
		`vim.o.background = "light"`, `  bg       = "#808a63", -- the ground`,
		`g(0, "Normal",        { fg = c.ink, bg = c.bg })`, "local term = {"} {
		if !strings.Contains(text, line+"\n") {
			t.Errorf("no %q in the scheme", line)
		}
	}
	if !strings.Contains(text, ":1  body prose") {
		t.Error("the ink has no contrast beside it")
	}
}

// a scheme this dialect wrote reads back to the same file, byte for byte.
func TestReadWriteIsIdentity(t *testing.T) {
	scheme, err := Of("test-dusk", colourway())
	if err != nil {
		t.Fatal(err)
	}
	first := written(t, scheme)
	again, err := Read(strings.NewReader(first))
	if err != nil {
		t.Fatal(err)
	}
	if again.Name != "test-dusk" {
		t.Errorf("the name read back as %q", again.Name)
	}
	if second := written(t, again); second != first {
		t.Error("read and written again, the scheme changed")
	}
}

// a scheme written by hand in the same shape reads too: a table with its
// own comments, groups in between, the terminal's table after.
func TestReadsAHandWrittenScheme(t *testing.T) {
	hand := `-- a scheme by hand
vim.g.colors_name = "by-hand"
local c = {
  bg      = "#808a63", -- the ground
  ink     = "#38241e", --  4.0:1   deep brown
}
g(0, "Normal", { fg = c.ink, bg = c.bg })
local term = {
  "#000000", "#110000", "#220000", "#330000", "#440000", "#550000", "#660000", "#770000",
  "#880000", "#990000", "#aa0000", "#bb0000", "#cc0000", "#dd0000", "#ee0000", "#ff0000",
}
`
	scheme, err := Read(strings.NewReader(hand))
	if err != nil {
		t.Fatal(err)
	}
	if got := ghostty.Hex(scheme.Colours["ink"]); got != "#38241e" {
		t.Errorf("ink read as %s", got)
	}
	if got := ghostty.Hex(scheme.Terminal[15]); got != "#ff0000" {
		t.Errorf("the sixteenth read as %s", got)
	}
	if scheme.Name != "by-hand" || !scheme.Light {
		t.Errorf("name %q, light %v", scheme.Name, scheme.Light)
	}
}

// a colourway missing a role is refused, by the role and what needs it.
func TestOfRefusesAMissingRole(t *testing.T) {
	roles := css.New()
	for _, rule := range colourway().Rules() {
		if rule.Name != "cursor-line" {
			roles.Set(rule.Name, rule.Swatch)
		}
	}
	if _, err := Of("x", roles); err == nil || !strings.Contains(err.Error(), "cursorln") {
		t.Errorf("got %v", err)
	}
}

// the roles a scheme says back make the same scheme.
func TestRolesRoundTrip(t *testing.T) {
	scheme, err := Of("test-dusk", colourway())
	if err != nil {
		t.Fatal(err)
	}
	back, err := Of("test-dusk", scheme.Roles())
	if err != nil {
		t.Fatal(err)
	}
	if written(t, back) != written(t, scheme) {
		t.Error("the scheme's own roles make a different scheme")
	}
}

// black on white is 21 to 1, and a colour on itself is 1 to 1.
func TestContrastEnds(t *testing.T) {
	black, white := srgb.MustHex("#000000").Swatch(), srgb.MustHex("#ffffff").Swatch()
	if got := Contrast(black, white); math.Abs(got-21) > 1e-6 {
		t.Errorf("black on white is %.6f", got)
	}
	if got := Contrast(white, white); math.Abs(got-1) > 1e-9 {
		t.Errorf("white on white is %.9f", got)
	}
}

// a colour off the lamps' byte grid is written as the file will hold
// it, so the contrast beside it, and the scheme read back, agree.
func TestOffTheGridReadsBackTheSame(t *testing.T) {
	roles := colourway()
	roles.Set("ink", srgb.RGB{R: 0.2201, G: 0.1403, B: 0.1187}.Swatch())
	scheme, err := Of("off-grid", roles)
	if err != nil {
		t.Fatal(err)
	}
	first := written(t, scheme)
	again, err := Read(strings.NewReader(first))
	if err != nil {
		t.Fatal(err)
	}
	if written(t, again) != first {
		t.Error("an off-grid colour read back differently")
	}
}

// a colour in a comment is not a colour in the table, and a file with
// no colour table is not a scheme.
func TestCommentsAreNotColours(t *testing.T) {
	hand := `vim.g.colors_name = "noted"
local c = {
  -- bg = "#ffffff", an old ground
  bg      = "#808a63", -- the ground, not "#000000"
  ink     = "#38241e",
}
local term = {
  -- "#ff0000" was red once
  "#000000", "#110000", "#220000", "#330000", "#440000", "#550000", "#660000", "#770000",
  "#880000", "#990000", "#aa0000", "#bb0000", "#cc0000", "#dd0000", "#ee0000", "#ff0000",
}
`
	scheme, err := Read(strings.NewReader(hand))
	if err != nil {
		t.Fatal(err)
	}
	if got := ghostty.Hex(scheme.Colours["bg"]); got != "#808a63" {
		t.Errorf("the ground read as %s", got)
	}
	if got := ghostty.Hex(scheme.Terminal[0]); got != "#000000" {
		t.Errorf("the first of the sixteen read as %s", got)
	}
	if _, err := Read(strings.NewReader("print('hello')\n")); err == nil {
		t.Error("a file with no table read as a scheme")
	}
}

// a scheme with no terminal table says no sixteen back, rather than
// sixteen blacks.
func TestNoSixteenSaysNone(t *testing.T) {
	scheme, err := Read(strings.NewReader("local c = {\n  bg = \"#808a63\",\n}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, found := scheme.Roles().Get("black"); found {
		t.Error("black was said for a scheme without the sixteen")
	}
}
