package main

import (
	"os"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/spaces/srgb"
)

// outOfGamut is a colour the screen cannot make: more chroma than the lamps
// give at that lightness. It is defined by this test.
const outOfGamut = "oklch(70% 0.4 30)"

// aFew is a count of steps for a ramp. It is defined by this test.
const aFew = 5

// verb is a verb of the command, set up to be run.
type verb func() error

// printing runs a verb with its output caught, and gives back the output and
// the verb's error. Unlike capture, a verb that fails does not stop the test.
func printing(t *testing.T, run verb) (string, error) {
	t.Helper()
	made := emptyFile(t)
	was := os.Stdout
	os.Stdout = made
	err := run()
	os.Stdout = was
	made.Close()
	return printed(t, made), err
}

// showing is a show of a colour.
type showing string

// run shows the colour.
func (colour showing) run() error { return show(string(colour)) }

// ramping is a ramp between two colours, in some steps.
type ramping struct {
	from, to string
	steps    int
}

// run draws the ramp.
func (between ramping) run() error {
	return ramp(between.from, between.to, between.steps)
}

// An oklch spelling with other than three numbers is refused, and says so.
// So is a hex code that is not one. The anti-test is the good spellings: they
// are read.
func TestParseRefusesASpellingItCannotRead(t *testing.T) {
	for _, bad := range []string{"oklch(1 2)", "oklch(1 2 3 4)", "oklch()",
		"#12", "nonsense", ""} {
		if _, err := parse(bad); err == nil {
			t.Errorf("%q was accepted", bad)
		}
	}
	_, err := parse("oklch(1 2)")
	if !strings.Contains(err.Error(), "three numbers") {
		t.Errorf("the refusal was %v", err)
	}
	for _, good := range []string{"oklch(50% 0.1 20)", "  #336699  "} {
		if _, err := parse(good); err != nil {
			t.Errorf("%q was refused: %v", good, err)
		}
	}
}

// show says whether the colour is in gamut. The anti-test is the colour that
// is not: it says it was clipped, and the one that is says nothing of it.
func TestShowSaysWhenAColourIsClipped(t *testing.T) {
	inside, err := printing(t, showing("#336699").run)
	if err != nil || !strings.Contains(inside, "in gamut") ||
		strings.Contains(inside, "clipped") {
		t.Errorf("an in-gamut colour: %q, %v", inside, err)
	}
	outside, err := printing(t, showing(outOfGamut).run)
	if err != nil || !strings.Contains(outside, "out of gamut, clipped") {
		t.Errorf("an out-of-gamut colour: %q, %v", outside, err)
	}
}

// ramp refuses a colour it cannot read, whichever of the two it is, and
// prints nothing. The anti-test is that two good colours print a ramp.
func TestRampRefusesAColourItCannotRead(t *testing.T) {
	for _, bad := range []ramping{{"nonsense", "#fff", aFew},
		{"#fff", "nonsense", aFew}} {
		out, err := printing(t, bad.run)
		if err == nil || out != "" {
			t.Errorf("ramp %v: %q, %v", bad, out, err)
		}
	}
	out, err := printing(t, ramping{"#000", "#fff", aFew}.run)
	if err != nil || out == "" {
		t.Errorf("a ramp of two good colours: %q, %v", out, err)
	}
}

// lowest is the fewest steps the property asks a ramp for. It is defined by
// this test.
const lowest = -50

// fewest is the fewest steps a ramp draws. It is defined by ramp.
const fewest = 2

// The property: a ramp asked for fewer than two steps draws two.
func TestRampDrawsAtLeastTwoSteps(t *testing.T) {
	want, _ := printing(t, ramping{"#000", "#fff", fewest}.run)
	rapid.Check(t, rampDraw{t, want}.holds)
}

// rampDraw is a test and what a ramp of the fewest steps drew.
type rampDraw struct {
	t    *testing.T
	want string
}

// holds is one draw of the property above.
func (each rampDraw) holds(draw *rapid.T) {
	steps := rapid.IntRange(lowest, fewest).Draw(draw, "steps")
	got, err := printing(each.t, ramping{"#000", "#fff", steps}.run)
	if err != nil || got != each.want {
		draw.Fatalf("%d steps drew differently from %d: %v", steps, fewest, err)
	}
}

// stdoutWas is the output a test replaced.
type stdoutWas struct{ file *os.File }

// put gives the output back.
func (was stdoutWas) put() { os.Stdout = was.file }

// cell is how wide a painted cell is in the test below. It is defined by
// this test.
const cell = 3

// Paint draws a cell when the output is a terminal and NO_COLOR is not set.
// A character device stands for a terminal, being the same kind of file. The
// anti-tests draw nothing: NO_COLOR is set, or the output is a plain file, or
// it cannot be examined.
func TestPaintDrawsOnATerminal(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	pink := srgb.MustHex("#ff6ec7").Swatch()
	device, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(stdoutWas{os.Stdout}.put)
	os.Stdout = device
	if got, want := paint(pink, cell), "\x1b[48;2;255;110;199m   \x1b[0m"; got != want {
		t.Errorf("on a terminal, paint drew %q, not %q", got, want)
	}
	t.Setenv("NO_COLOR", "1")
	if got := paint(pink, cell); got != "" {
		t.Errorf("under NO_COLOR, paint drew %q", got)
	}
	t.Setenv("NO_COLOR", "")
	device.Close()
	if got := paint(pink, cell); got != "" {
		t.Errorf("on output that cannot be examined, paint drew %q", got)
	}
	os.Stdout = emptyFile(t)
	if got := paint(pink, cell); got != "" {
		t.Errorf("on a plain file, paint drew %q", got)
	}
}
