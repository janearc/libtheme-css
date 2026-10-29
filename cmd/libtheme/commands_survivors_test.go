package main

import (
	"math"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/radiation"
)

// These tests hold what the mutation run found nothing holding. They cover
// the width of each painted cell and the count in each bar. They cover the
// exit statuses as numbers and the spellings parse takes.

const (
	// shellDone is the status of a call that worked. It is defined by the
	// shell.
	shellDone = 0
	// shellFailed is the status of a call that failed. It is defined by the
	// shell.
	shellFailed = 1
	// shellWrong is the status of a call that makes no sense. It is defined
	// by the shell.
	shellWrong = 2
)

const (
	// showWidth is the width of the one cell show paints. It is defined by
	// show.
	showWidth = 12
	// lineWidth is the width of a cell in a line or a bar. It is defined by
	// the verbs that paint them.
	lineWidth = 2
	// swatchWidth is the width of the cell before each colour known lists.
	// It is defined by known.
	swatchWidth = 8
	// stopWidth is the width of the cell before each stop read lists. It is
	// defined by read.
	stopWidth = 4
	// sampleWidth is the width of a cell in the bar read draws for a ramp.
	// It is defined by read.
	sampleWidth = 1
	// readSamples is how many cells that bar has. It is defined by read.
	readSamples = 48
	// knownColours is how many colours known and the round trip list. It is
	// defined by the lists in main.go.
	knownColours = 5
	// mixerLines is how many lines a ramp draws. It is one in oklab and one
	// in srgb. It is defined by the two mixers.
	mixerLines = 2
	// barCells is how many cells a bar of radiation has. It is defined by
	// the usage line.
	barCells = 24
)

const (
	// belowTheLeast is a count of steps below what a ramp draws. It is
	// defined by this test.
	belowTheLeast = -3
	// justOne is one step. It is defined by counting.
	justOne = 1
	// theLeast is the fewest cells a ramp draws. It is defined by ramp.
	theLeast = 2
	// threeSteps is a count of steps just above the least. It is defined
	// by this test.
	threeSteps = 3
	// manySteps is a count of steps well above it. It is defined by this
	// test.
	manySteps = 7
	// half is a half. It is defined by halving.
	half = 0.5
	// allMatches asks regexp for every match. It is defined by regexp.
	allMatches = -1
	// towardsInfinity is the sign math.Inf is given for the positive
	// infinity. It is defined by math.Inf.
	towardsInfinity = 1
)

// cellSpaces is the group of the pattern below that holds the cell's spaces.
const cellSpaces = "spaces"

// cellRe is one painted cell. It has a background escape, spaces and a reset.
var cellRe = regexp.MustCompile(
	"\x1b\\[48;2;\\d+;\\d+;\\d+m(?P<spaces> *)\x1b\\[0m")

// paintsAlways is the answer of a terminal.
func paintsAlways() bool { return true }

// restoring puts painting back as it was.
type restoring struct{ was func() bool }

// put is the restoring.
func (before restoring) put() { painting = before.was }

// drawing runs a verb as if the output were a terminal. It gives back the
// width of every cell painted, in order. It gives back the text too.
func drawing(t *testing.T, run verb) (widths []int, text string) {
	t.Helper()
	t.Cleanup(restoring{painting}.put)
	painting = paintsAlways
	text, err := printing(t, run)
	if err != nil {
		t.Fatal(err)
	}
	at := cellRe.SubexpIndex(cellSpaces)
	for _, cell := range cellRe.FindAllStringSubmatch(text, allMatches) {
		widths = append(widths, len(cell[at]))
	}
	return widths, text
}

// counted is how many cells of a width there are.
func counted(widths []int, width int) int {
	count := nothing
	for _, each := range widths {
		if each == width {
			count++
		}
	}
	return count
}

// calling is a call of the verb table.
type calling []string

// run dispatches the call.
func (words calling) run() error { return dispatch(words) }

// sheetPainting is a paint of a sheet.
type sheetPainting struct{ sheet *css.Sheet }

// run paints the sheet.
func (draw sheetPainting) run() error {
	paintSheet(draw.sheet)
	return nil
}

// statusCall is a call and the status it stops with.
type statusCall struct {
	status    int
	arguments []string
}

// statusCalls are one call that works, one that fails and one that makes no
// sense.
var statusCalls = []statusCall{
	{shellDone, []string{"version"}},
	{shellFailed, []string{"show", "nonsense"}},
	{shellWrong, []string{"nosuch"}},
}

// The statuses are the shell's numbers. Each call stops with its own. That
// is the anti-test.
func TestMainStopsWithTheShellsNumbers(t *testing.T) {
	for _, each := range statusCalls {
		made := asProcess(t, each.arguments...)
		made.stopper.EXPECT().Exit(each.status)
		main()
	}
}

// Show paints one cell, twelve wide.
func TestShowPaintsOneCellTwelveWide(t *testing.T) {
	widths, _ := drawing(t, showing("#336699").run)
	if len(widths) != justOne || counted(widths, showWidth) != justOne {
		t.Errorf("show painted %v", widths)
	}
}

// asking is a count of steps for a ramp and the cells a line of it has.
type asking struct{ steps, cells int }

// askings are counts below, at and above the least a ramp draws.
var askings = []asking{
	{belowTheLeast, theLeast}, {nothing, theLeast}, {justOne, theLeast},
	{theLeast, theLeast}, {threeSteps, threeSteps}, {manySteps, manySteps},
}

// Ramp draws two lines of the cells it is asked for. Fewer than two gives
// two. Counts above two are not raised, as the anti-test.
func TestRampDrawsTheCellsItIsAskedFor(t *testing.T) {
	for _, each := range askings {
		widths, _ := drawing(t, ramping{"#ff0000", "#0000ff", each.steps}.run)
		if len(widths) != mixerLines*each.cells ||
			counted(widths, lineWidth) != len(widths) {
			t.Errorf("asked %d: painted %v, wanted %d cells of %d twice",
				each.steps, widths, each.cells, lineWidth)
		}
	}
}

// Ramp with no count draws the default, which is the usage line's 24.
func TestRampWithNoCountDrawsTwentyFourCells(t *testing.T) {
	widths, _ := drawing(t, calling{"ramp", "#ff0000", "#0000ff"}.run)
	if counted(widths, lineWidth) != mixerLines*barCells {
		t.Errorf("default ramp painted %d cells", counted(widths, lineWidth))
	}
}

// Ramp starts at the first colour it is given. It ends at the second.
func TestRampRunsFromTheFirstColourToTheSecond(t *testing.T) {
	_, text := drawing(t, calling{"ramp", "#ff0000", "#0000ff"}.run)
	if !strings.Contains(text, "#ff0000 0%, #0000ff 100%") {
		t.Errorf("the ramp is not red to blue:\n%s", text)
	}
}

// Known draws a swatch before each colour. It draws two lines of cells and
// a bar for each profile. The sheet mode draws a cell for each colour. The
// css mode draws none.
func TestKnownDrawsItsCells(t *testing.T) {
	widths, _ := drawing(t, calling{"known"}.run)
	bars := mixerLines + len(radiation.Profiles)
	if counted(widths, swatchWidth) != knownColours ||
		counted(widths, lineWidth) != bars*barCells ||
		len(widths) != knownColours+bars*barCells {
		t.Errorf("known painted %d swatches and %d cells in %d",
			counted(widths, swatchWidth), counted(widths, lineWidth),
			len(widths))
	}
	widths, _ = drawing(t, calling{"known", "--paint"}.run)
	if len(widths) != knownColours ||
		counted(widths, lineWidth) != knownColours {
		t.Errorf("the sheet painted %v", widths)
	}
	widths, text := drawing(t, calling{"known", "--css"}.run)
	if len(widths) != nothing || !strings.HasPrefix(text, ":root {") {
		t.Errorf("--css gave %q with cells %v", text, widths)
	}
}

// The round trip draws a cell for each colour it tries, and a bar for each
// range of radiation.
func TestRoundtripDrawsItsCells(t *testing.T) {
	widths, _ := drawing(t, calling{"roundtrip"}.run)
	bars := len(byRender(radiation.Profiles))
	if counted(widths, lineWidth) != len(widths) ||
		len(widths) != knownColours+bars*barCells {
		t.Errorf("the round trip painted %v", widths)
	}
}

// Read draws each stop four wide and each ramp as a bar of 48 cells, one
// wide.
func TestReadDrawsItsCells(t *testing.T) {
	widths, _ := drawing(t, calling{"read", "testdata/dusk.css"}.run)
	src, err := os.ReadFile("testdata/dusk.css")
	if err != nil {
		t.Fatal(err)
	}
	stops, ramps := nothing, nothing
	cw := css.Read(string(src))
	for _, name := range cw.Order {
		stops += len(cw.Ramps[name].Stops)
		ramps++
	}
	if ramps == nothing || counted(widths, stopWidth) != stops ||
		counted(widths, sampleWidth) != readSamples*ramps {
		t.Errorf("read painted %d stops for %d and %d samples for %d ramps",
			counted(widths, stopWidth), stops, counted(widths, sampleWidth),
			ramps)
	}
}

// gammaTrip is gamma's bar on the trip, a cell at a time. The x cells do
// not come back. The stops do. It was read from the round trip's printout.
const gammaTrip = "xxxxx....xxx....xxxxxxxx"

// marksOf is the row of marks for a trip: two characters a cell, an x for a
// cell that did not come back and a blank for one that did.
func marksOf(trip string) string {
	var marks strings.Builder
	for _, each := range trip {
		if each == 'x' {
			marks.WriteString(" x")
			continue
		}
		marks.WriteString("  ")
	}
	return marks.String()
}

// Cells that do not come back are marked with an x. The anti-test is that
// the marks and the count agree.
func TestTripAcrossMarksTheCellsThatDoNotComeBack(t *testing.T) {
	bar, marks, back := tripAcross(radiation.Gamma)
	if marks != marksOf(gammaTrip) || back != strings.Count(gammaTrip, ".") ||
		bar != "" {
		t.Errorf("marks %q back %d bar %q", marks, back, bar)
	}
}

// reading is a distance and what the round trip says of it.
type reading struct {
	distance float64
	verdict  string
}

// The verdicts at each bound.
const (
	same      = "identical"
	alike     = "same to a person"
	different = "different to a person"
)

// Exact or less is identical. Past Eye is a loss. Eye itself is not.
// Between them is the same to a person.
func TestVerdictOfDrawsItsLinesAtTheBounds(t *testing.T) {
	for _, each := range []reading{
		{nothing, same}, {ok.Exact, same},
		{math.Nextafter(ok.Exact, ok.Eye), alike}, {ok.Eye, alike},
		{math.Nextafter(ok.Eye, math.Inf(towardsInfinity)), different},
	} {
		got, lost := verdictOf(each.distance)
		if got != each.verdict || lost != (each.verdict == different) {
			t.Errorf("%g: %q lost %v, wanted %q", each.distance, got, lost,
				each.verdict)
		}
	}
}

// outside is the cell of a wavelength outside a profile's source. It is
// defined by cellOf.
const outside = -1

// wavelengthCell is a wavelength and the cell it falls in.
type wavelengthCell struct {
	nm   float64
	cell int
}

// The start of a source is in the first cell. Its end is out. Before the
// start is out. Just inside the end is the last cell.
func TestCellOfHoldsTheEdges(t *testing.T) {
	source := radiation.Gamma.Source
	for _, each := range []wavelengthCell{
		{source.Lo, nothing}, {source.Hi, outside}, {source.Lo * half, outside},
		{math.Nextafter(source.Hi, source.Lo), cells - justOne},
	} {
		if got := cellOf(radiation.Gamma, each.nm); got != each.cell {
			t.Errorf("%g: cell %d, wanted %d", each.nm, got, each.cell)
		}
	}
}

// samePerson is whether two spellings are the same colour to arithmetic.
func samePerson(t *testing.T, one, other string) bool {
	t.Helper()
	first, err := parse(one)
	if err != nil {
		t.Fatalf("%q: %v", one, err)
	}
	second, err := parse(other)
	if err != nil {
		t.Fatalf("%q: %v", other, err)
	}
	apart := ok.Distance(ok.FromSwatch(first), ok.FromSwatch(second))
	return apart <= ok.Exact
}

// Percent means a hundredth. The hue none means zero. Pairs that differ
// are the anti-test.
func TestParseReadsPercentAndNone(t *testing.T) {
	if !samePerson(t, "oklch(50% 0.1 20)", "oklch(0.5 0.1 20)") {
		t.Error("50% is not 0.5")
	}
	if samePerson(t, "oklch(51% 0.1 20)", "oklch(0.5 0.1 20)") {
		t.Error("51% is 0.5")
	}
	if !samePerson(t, "oklch(50% 0.1 none)", "oklch(50% 0.1 0)") {
		t.Error("none is not a hue of zero")
	}
	if samePerson(t, "oklch(50% 0.1 none)", "oklch(50% 0.1 1)") {
		t.Error("none is a hue of one")
	}
}

// The text oklch( with no closing bracket is not read. The anti-test is the
// same spelling closed.
func TestParseNeedsTheClosingBracket(t *testing.T) {
	if _, err := parse("oklch(50% 0.1 20"); err == nil {
		t.Error("an open oklch( was read")
	}
	if _, err := parse("oklch(50% 0.1 20)"); err != nil {
		t.Error(err)
	}
}

// The values of a sheet line up on the longest name, and no further: a sheet
// of one nameless rule pads nothing.
func TestPaintSheetPadsToTheLongestName(t *testing.T) {
	sheet := css.New()
	sheet.Set("", swatch.NoLight())
	text, _ := printing(t, sheetPainting{sheet}.run)
	if !strings.Contains(text, "--: #000000;") {
		t.Errorf("a nameless rule was padded: %q", text)
	}
	sheet.Set("ab", swatch.NoLight())
	text, _ = printing(t, sheetPainting{sheet}.run)
	if !strings.Contains(text, "--:   #000000;") ||
		!strings.Contains(text, "--ab: #000000;") {
		t.Errorf("the values did not line up: %q", text)
	}
}
