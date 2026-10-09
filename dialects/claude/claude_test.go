package claude

import (
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// a colourway's claude colours come out by the numbers claude code draws
// in, the band among them as the terminal's black.
func TestPaletteIsByNumber(t *testing.T) {
	roles := css.New()
	roles.Set("claude-text", srgb.MustHex("#c5e699").Swatch())
	roles.Set("black", srgb.MustHex("#57696a").Swatch())
	palette := Palette(roles)
	if got := ghostty.Hex(palette[252]); got != "#c5e699" {
		t.Errorf("252 is %s", got)
	}
	if got := ghostty.Hex(palette[0]); got != "#57696a" {
		t.Errorf("the band is %s", got)
	}
	if len(palette) != 2 {
		t.Errorf("%d colours from two roles", len(palette))
	}
}

// Into says the names as numbers for the terminal theme, keeps
// everything else, and refuses a colour set under both.
func TestIntoSaysTheNumbers(t *testing.T) {
	roles := css.New()
	roles.Set("ink", srgb.MustHex("#38241e").Swatch())
	roles.Set("claude-spinner", srgb.MustHex("#ffaa37").Swatch())
	said, err := Into(roles)
	if err != nil {
		t.Fatal(err)
	}
	if value, found := said.Get("palette-219"); !found || ghostty.Hex(value) != "#ffaa37" {
		t.Error("the spinner is not palette-219")
	}
	if _, found := said.Get("ink"); !found {
		t.Error("the ink was lost")
	}
	if _, found := roles.Get("palette-219"); found {
		t.Error("Into changed the colourway it was given")
	}
	roles.Set("palette-219", srgb.MustHex("#000000").Swatch())
	if _, err := Into(roles); err == nil || !strings.Contains(err.Error(), "claude-spinner") {
		t.Errorf("both set: %v", err)
	}
}

// a palette read from a theme names claude's colours back, and leaves
// the band to the terminal's black.
func TestOfNamesThePalette(t *testing.T) {
	palette := map[int]swatch.Swatch{252: srgb.MustHex("#c5e699").Swatch(),
		0: srgb.MustHex("#57696a").Swatch(), 22: srgb.MustHex("#c19a43").Swatch()}
	roles := Of(palette)
	if names := strings.Join(roles.Names(), " "); names != "claude-text claude-added" {
		t.Errorf("named %q", names)
	}
}

// the band, set as black, stays the terminal's: Into says no palette-0,
// which the ghostty dialect would refuse.
func TestIntoLeavesTheBandToBlack(t *testing.T) {
	roles := css.New()
	roles.Set("black", srgb.MustHex("#57696a").Swatch())
	said, err := Into(roles)
	if err != nil {
		t.Fatal(err)
	}
	if _, found := said.Get("palette-0"); found {
		t.Error("the band was said as palette-0")
	}
}
