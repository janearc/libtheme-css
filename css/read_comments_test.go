package css_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// nothing is none of anything. It is defined by counting.
const nothing = 0

// half is the middle of a range. It is defined by halving.
const half = 2

// black is the colour a test checks a role did not change from.
var black = srgb.MustHex("#000000").Swatch()

// roleIs is whether the role is there and is the swatch.
func roleIs(sheet *css.Sheet, name string, want srgb.RGB) bool {
	got, has := sheet.Get(name)
	return has && got == want.Swatch()
}

// A declaration in a comment is not a role, whether the comment holds it
// whole, comes after the live one, or never ends. The live one is read.
func TestCommentsAreNotRoles(t *testing.T) {
	cases := map[string]struct {
		source string
		live   bool
	}{
		"whole":           {"/* --ink: #000000; */", false},
		"never ends":      {":root { --x: #ffffff; } /* --ink: #000000;", false},
		"after the live":  {":root { --ink: #ffffff; /* --ink: #000000; */ }", true},
		"before the live": {"/* --ink: #000000; */ :root { --ink: #ffffff; }", true},
		"over lines":      {"/*\n --ink: #000000;\n*/ :root { --ink: #ffffff; }", true},
		"in the value":    {":root { --ink: /* the paper */ #ffffff; }", true},
		"in the name":     {":root { --ink /* x */ : #ffffff; }", true},
	}
	for name, one := range cases {
		roles := css.Read(one.source).Roles
		_, has := roles.Get("ink")
		if has != one.live || (one.live && !roleIs(roles, "ink", srgb.MustHex("#ffffff"))) {
			t.Errorf("%s: ink read %v, want %v", name, has, one.live)
		}
	}
}

// Declarations outside any comment are still read after a comment that
// closed, and a comment mark inside the file after it does not hide them.
func TestAClosedCommentHidesOnlyItself(t *testing.T) {
	roles := css.Read("/* a */ :root { --a: #ffffff; } /* b */ " +
		":root { --b: #000000; } /* c */").Roles
	if !roleIs(roles, "a", srgb.MustHex("#ffffff")) || !roleIs(roles, "b", srgb.MustHex("#000000")) {
		t.Errorf("roles read: %v", roles.Names())
	}
}

// A value that is not a complete colour is not read, and is not lost
// silently: it is listed with its line.
func TestAnIncompleteColourIsNotRead(t *testing.T) {
	source := ":root {\n  --a: #fff.foo;\n  --b: #12;\n  --c: #12345;\n" +
		"  --d: #ff00ff80;\n  --e: #gggggg;\n  --f: #ffffffx;\n}\n"
	read := css.Read(source)
	if names := read.Roles.Names(); len(names) != nothing {
		t.Errorf("roles read from incomplete colours: %v", names)
	}
	skipped := strings.Join(read.Skipped(), "\n")
	for _, want := range []string{"line 2: --a: #fff.foo", "line 3: --b: #12",
		"line 4: --c: #12345", "line 5: --d: #ff00ff80", "line 6: --e: #gggggg",
		"line 7: --f: #ffffffx"} {
		if !strings.Contains(skipped, want) {
			t.Errorf("no %q in\n%s", want, skipped)
		}
	}
}

// Complete colours are read whatever follows them: the end of the rule, a
// comment, or !important, and 3 or 6 digits in either case.
func TestCompleteColoursAreRead(t *testing.T) {
	source := ":root { --a: #FFF; --b: #ffffff!important; --c: #ffffff /* x */;\n" +
		"  --d: #FFFFFF }"
	read := css.Read(source)
	for _, name := range []string{"a", "b", "c", "d"} {
		if !roleIs(read.Roles, name, srgb.MustHex("#ffffff")) {
			t.Errorf("%s was not read as white", name)
		}
	}
	if len(read.Skipped()) != nothing {
		t.Errorf("complete colours were skipped: %v", read.Skipped())
	}
}

// What is not a colour value is none of the reader's business: nothing is
// read, and nothing is listed as skipped.
func TestOtherPropertiesAreLeftAlone(t *testing.T) {
	source := ":root { --font: Helvetica; --shadow: 0 1px #000; " +
		"--anchor: url(#top); --space: 4px; --name: \"#hash\"; }"
	read := css.Read(source)
	if len(read.Roles.Names()) != nothing || len(read.Skipped()) != nothing {
		t.Errorf("read %v, skipped %v", read.Roles.Names(), read.Skipped())
	}
}

// A role set twice is the later one, as css has it, and keeps its place.
func TestALaterDeclarationWins(t *testing.T) {
	read := css.Read(":root { --a: #000000; --b: #000000; --a: #ffffff; }")
	if !roleIs(read.Roles, "a", srgb.MustHex("#ffffff")) || !slices.Equal(read.Roles.Names(), []string{"a", "b"}) {
		t.Errorf("roles: %v", read.Roles.Names())
	}
	if got, _ := read.Roles.Get("b"); got != black {
		t.Error("b changed")
	}
}

// A family of one is a role and no ramp: a ramp needs two stops.
func TestAFamilyOfOneIsNotARamp(t *testing.T) {
	read := css.Read(":root { --sunset-1: #000000; --dusk-1: #000000; --dusk-2: #ffffff; }")
	if _, has := read.Ramps["sunset"]; has || !slices.Equal(read.Order, []string{"dusk"}) {
		t.Errorf("ramps: %v", read.Order)
	}
}

// hexOf is a swatch as the hex a sheet would write it.
func hexOf(colour swatch.Swatch) string {
	rgb, _ := srgb.FromSwatch(colour)
	return rgb.Hex()
}

// A ramp is its family in the order of their numbers, not the file's, at
// even stops, mixed in oklab. Families come in the order the file first
// names them. Two members with the same number stay in the file's order. A
// role declared twice is its later colour, in its first place.
func TestRampsFollowTheirNumbers(t *testing.T) {
	source := ":root { --z-2: #000000; --z-1: #ffffff; --z-3: #808080;\n" +
		"--solo-1: #123456; --tie-1: #111111; --tie-01: #222222;\n" +
		"--y-1: #000000; --y-2: #ffffff; --y-1: #808080; }"
	read := css.Read(source)
	if want := []string{"z", "tie", "y"}; !slices.Equal(read.Order, want) {
		t.Errorf("ramps in order %v, want %v", read.Order, want)
	}
	for name, want := range map[string]string{
		"z":   "linear-gradient(in oklab, #ffffff 0%, #000000 50%, #808080 100%)",
		"tie": "linear-gradient(in oklab, #111111 0%, #222222 100%)",
		"y":   "linear-gradient(in oklab, #808080 0%, #ffffff 100%)",
	} {
		if got := read.Ramps[name].String(hexOf); got != want {
			t.Errorf("%s: %s, want %s", name, got, want)
		}
	}
	if _, has := read.Ramps["solo"]; has || len(read.Ramps) != len(read.Order) {
		t.Errorf("ramps: %v", read.Ramps)
	}
}

// A colour the lamps cannot make is written as the nearest they can, and the
// comment beside it says it was clipped.
func TestAnOutOfGamutColourSaysClipped(t *testing.T) {
	sheet := css.New()
	low, high := swatch.Observed()
	sheet.Set("green", swatch.Monochrome(low+(high-low)/half))
	if text := sheet.String(); !strings.Contains(text, ", clipped */") {
		t.Errorf("an out-of-gamut colour was not marked:\n%s", text)
	}
}
