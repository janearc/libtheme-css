package readable_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/readable"
)

// picked is the numbers these tests pick, from testdata/ranges.json.
type picked struct {
	Refused map[string]json.RawMessage
	Newer   json.RawMessage
	Band    struct{ Low, High float64 }       `json:"reference band"`
	Draws   struct{ Lowest, Highest float64 } `json:"nesting draws"`
	Rules   struct {
		Ratio  float64 `json:"minimum ratio"`
		Circle float64 `json:"half circle"`
	}
}

// pickedNumbers reads the test numbers, or the test stops.
func pickedNumbers(t interface{ Fatal(...any) }) picked {
	raw, err := os.ReadFile("testdata/ranges.json")
	if err != nil {
		t.Fatal(err)
	}
	var numbers picked
	if err := json.Unmarshal(raw, &numbers); err != nil {
		t.Fatal(err)
	}
	return numbers
}

// nothing and whole are the ends of the saturation scale. They are defined
// by the scale.
const (
	// nothing is no saturation. It is defined by the scale.
	nothing = 0
	// whole is all the saturation there is. It is defined by the scale.
	whole = 1
)

// readText is a range read from text.
func readText(text string) (readable.Policy, error) {
	return readable.Read(strings.NewReader(text))
}

// The reference range is the one measured on 2026-09-03: prose inside the
// band is inside it, whatever the grey, and outside is named.
func TestTheReferenceBandIsTheMeasuredOne(t *testing.T) {
	rapid.Check(t, bandAgrees)
}

// bandAgrees is one draw of the property above: a grey ink on a black
// ground is outside the band exactly when its contrast is.
func bandAgrees(t *rapid.T) {
	band := pickedNumbers(t).Band
	level := rapid.Uint8().Draw(t, "level")
	grey := fmt.Sprintf("#%02x%02x%02x", level, level, level)
	black := map[string]string{"bg": "#000000", "cursorln": "#000000",
		"visual": "#000000", "panel": "#000000", "ink": grey}
	scheme := schemeOf(changed(black))
	got := nvim.Contrast(scheme.Colours["ink"], scheme.Colours["bg"])
	policy, err := readable.Reference()
	if err != nil {
		t.Fatal(err)
	}
	report := readable.Check(scheme, policy)
	found := strings.Contains(joined(report), "ink on the ground")
	if found != (got < band.Low || got > band.High) {
		t.Fatalf("grey %s is %.3f:1 and the gate said %v", grey, got, found)
	}
}

// The anti-tests: a range whose pieces do not nest, or that is not a
// number, or a file that is not a range, is refused.
func TestARangeThatMakesNoSenseIsRefused(t *testing.T) {
	for name, document := range pickedNumbers(t).Refused {
		if _, err := readText(string(document)); err == nil {
			t.Errorf("%s was read as a range", name)
		}
	}
	for name, text := range map[string]string{"not json": "{",
		"too big a number": `{"contrast": {"low": 1e999}}`} {
		if _, err := readText(text); err == nil {
			t.Errorf("%s was read as a range", name)
		}
	}
}

// A range from a newer measurement is read, and held to.
func TestANewerRangeIsRead(t *testing.T) {
	newer := string(pickedNumbers(t).Newer)
	policy, err := readText(newer)
	if err != nil {
		t.Fatal(err)
	}
	if readable.Check(schemeOf(inBand), policy).OK() {
		t.Error("a scheme at 11:1 passed a band that ends at 7:1")
	}
}

// The reference file says where it came from and when.
func TestTheReferenceRangeSaysWhereItCameFrom(t *testing.T) {
	raw, err := os.ReadFile("reference-dark-2026-09-03.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"2026-09-03", "libreadme"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("the reference range does not say %q", want)
		}
	}
}

// The property: a range is read exactly when its numbers nest, whatever
// they are.
func TestARangeIsReadExactlyWhenItNests(t *testing.T) {
	rapid.Check(t, nestsAsExpected)
}

// nestsAsExpected is one draw of the property above.
func nestsAsExpected(t *rapid.T) {
	numbers := pickedNumbers(t)
	number := rapid.Float64Range(numbers.Draws.Lowest, numbers.Draws.Highest)
	low, high := number.Draw(t, "low"), number.Draw(t, "high")
	floor, apart := number.Draw(t, "floor"), number.Draw(t, "apart")
	neutral := number.Draw(t, "neutral")
	want := floor > numbers.Rules.Ratio && floor <= low && low < high &&
		apart > nothing && apart <= numbers.Rules.Circle &&
		neutral > nothing && neutral < whole
	text := fmt.Sprintf(`{"contrast": {"low": %g, "high": %g, "state_min": %g},
		"hue": {"min_separation_degrees": %g, "neutral_saturation": %g}}`,
		low, high, floor, apart, neutral)
	if _, err := readText(text); (err == nil) != want {
		t.Fatalf("range %s: error %v, but nesting is %v", text, err, want)
	}
}
