package readable_test

import (
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/readable"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// inBand is a scheme every prose colour of which is inside the band on
// every surface, every state colour above the floor, and every pair a
// group draws above the floor. The grounds are near black greys; the inks
// are greys that sit in the band on all four grounds.
var inBand = map[string]string{
	"bg": "#0a0a0a", "cursorln": "#101010", "visual": "#161616",
	"panel": "#0e0e0e", "ink": "#c8c8c8", "dim": "#c4c4c4",
	"heading": "#c6c6c6", "link": "#c5c5c5", "code": "#c7c7c7",
	"string": "#c9c9c9", "gutter": "#4a4a4a", "border": "#3a3a3a",
	"err": "#e07070", "warn": "#d9a050", "info": "#70a0e0",
	"hint": "#70d0b0",
}

// schemeOf is a scheme of the colours given by name.
func schemeOf(colours map[string]string) nvim.Scheme {
	var scheme nvim.Scheme
	scheme.Name = "test"
	scheme.Colours = map[string]swatch.Swatch{}
	for name, hex := range colours {
		scheme.Colours[name] = srgb.MustHex(hex).Swatch()
	}
	return scheme
}

// changed is a copy of inBand with some colours set to other hexes.
func changed(replacements map[string]string) map[string]string {
	colours := map[string]string{}
	for name, hex := range inBand {
		colours[name] = hex
	}
	for name, hex := range replacements {
		colours[name] = hex
	}
	return colours
}

// joined is a report's lines as one text.
func joined(report readable.Report) string {
	return strings.Join(report.Lines(), "\n")
}

// reference is the reader's range as measured, or the test stops.
func reference(t *testing.T) readable.Policy {
	t.Helper()
	policy, err := readable.Reference()
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

// checked is the scheme made of the colours, held to the reference range,
// as its report.
func checked(t *testing.T, colours map[string]string) readable.Report {
	t.Helper()
	return readable.Check(schemeOf(colours), reference(t))
}

// A scheme inside the range everywhere passes. The anti-tests below need
// this: a gate that fails everything would pass them.
func TestASchemeInsideTheRangePasses(t *testing.T) {
	report := checked(t, inBand)
	if !report.OK() {
		t.Errorf("a scheme in range failed:\n%s",
			joined(report))
	}
}

// The anti-tests: each colour that leaves the range is named, with the
// surface it fails on and the rule it breaks.
func TestEachWayOutOfTheRangeIsNamed(t *testing.T) {
	cases := map[string]struct {
		replace map[string]string
		want    string
	}{
		"ink too dim": {map[string]string{"ink": "#505050"},
			"ink on the ground: too dim to read at length"},
		"comments glare": {map[string]string{"dim": "#ffffff"},
			"dim on the ground: glare"},
		"state too dim": {map[string]string{"err": "#301010"},
			"err on the ground: too dim to scan"},
		"on the selection": {map[string]string{"visual": "#3a3a3a"},
			"link on the selection: too dim to read at length"},
		"on a panel": {map[string]string{"panel": "#303030"},
			"link on the panel: too dim to read at length"},
		"on the line": {map[string]string{"cursorln": "#303030"},
			"code on the cursor line: too dim to read at length"},
		"no colour edge": {map[string]string{"bg": "#14507a",
			"heading": "#2f7ac0"}, "heading on the ground: no colour edge"},
		"search match": {map[string]string{"string": "#101010"},
			"Search draws bg on string: too dim to scan"},
	}
	for name, one := range cases {
		report := checked(t, changed(one.replace))
		lines := joined(report)
		if report.OK() || !strings.Contains(lines, one.want) {
			t.Errorf("%s: wanted %q in\n%s", name, one.want, lines)
		}
	}
}

// Surfaces that are the same hue as the text but desaturated pass the hue
// rule: the ink's own colour carries the edge.
func TestANeutralSurfaceNeedsNoHueGap(t *testing.T) {
	tinted := changed(map[string]string{"bg": "#0b0a0a", "ink": "#c8c0c0"})
	if report := checked(t, tinted); !report.OK() {
		t.Errorf("a neutral surface failed:\n%s",
			joined(report))
	}
}

// Furniture is left alone: a gutter and a border as dim as you like pass.
func TestFurnitureHasNoRule(t *testing.T) {
	dim := changed(map[string]string{"gutter": "#0c0c0c", "border": "#0c0c0c"})
	if report := checked(t, dim); !report.OK() {
		t.Errorf("furniture was held to a rule:\n%s",
			joined(report))
	}
}

// A colour the scheme lacks is a failure, not a skip, and a scheme with
// nothing in it is not a pass.
func TestAMissingColourFails(t *testing.T) {
	scheme := schemeOf(inBand)
	delete(scheme.Colours, "panel")
	report := readable.Check(scheme, reference(t))
	lines := joined(report)
	if report.OK() || !strings.Contains(lines, "panel: missing colour") {
		t.Errorf("a scheme without a panel gave:\n%s", lines)
	}
	var empty nvim.Scheme
	if readable.Check(empty, reference(t)).OK() {
		t.Error("a scheme with nothing in it passed")
	}
}

// A failure says what is wrong in one line, with the numbers.
func TestAFailureReadsAsAnAction(t *testing.T) {
	report := checked(t, changed(map[string]string{"ink": "#505050"}))
	first := joined(report)
	for _, want := range []string{
		"test: ink on the ground: too dim to read at length",
		"#505050", "the band starts at 10:1"} {
		if !strings.Contains(first, want) {
			t.Errorf("no %q in %q", want, first)
		}
	}
}

// The lines are a copy: changing what Lines returns does not change the
// report.
func TestLinesAreACopy(t *testing.T) {
	report := checked(t, changed(map[string]string{"ink": "#505050"}))
	before := joined(report)
	copied := report.Lines()
	clear(copied)
	if joined(report) != before {
		t.Error("the report changed when its copy did")
	}
}
