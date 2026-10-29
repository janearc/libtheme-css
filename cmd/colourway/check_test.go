package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// bandSource is a source every colour of which is inside the reference
// range on every surface. The text is what a person would write.
const bandSource = `:root {
  --ground: #0a0a0a; --cursor-line: #101010; --surface-1: #161616;
  --panel: #0e0e0e;
  --ink: #c8c8c8; --dim: #c4c4c4; --heading: #c6c6c6; --link: #c5c5c5;
  --code: #c7c7c7; --string: #c9c9c9; --gutter: #4a4a4a; --border: #3a3a3a;
  --err: #e07070; --warn: #d9a050; --info: #70a0e0; --hint: #70d0b0;
  --black: #101010; --red: #aa5555; --green: #55aa55; --yellow: #aaaa55;
  --blue: #5555aa; --magenta: #aa55aa; --cyan: #55aaaa; --white: #aaaaaa;
  --bright-black: #555555; --bright-red: #ff5555; --bright-green: #55ff55;
  --bright-yellow: #ffff55; --bright-blue: #5555ff;
  --bright-magenta: #ff55ff; --bright-cyan: #55ffff; --bright-white: #ffffff;
}
`

// newerRange is a measured range the band source is outside: its band ends
// at 7:1.
const newerRange = `{"contrast": {"low": 4.5, "high": 7, "state_min": 3},
  "hue": {"min_separation_degrees": 30, "neutral_saturation": 0.2}}`

// checked runs check on the arguments and returns what it printed.
func checked(t *testing.T, arguments ...string) (printed string, err error) {
	t.Helper()
	var out bytes.Buffer
	err = check(arguments, streams{out: &out})
	return out.String(), err
}

func TestACheckedSourceInsideTheRangePasses(t *testing.T) {
	printed, err := checked(t, stored(t, "band.css", bandSource))
	if printed != "" || err != nil {
		t.Errorf("a source in range gave %q, %v", printed, err)
	}
}

// The anti-test: a source outside the range fails and says which colour on
// which surface, and a second source is still checked after the first.
func TestACheckedSourceOutsideTheRangeFails(t *testing.T) {
	dim := strings.ReplaceAll(bandSource, "--ink: #c8c8c8", "--ink: #505050")
	printed, err := checked(t, stored(t, "dim.css", dim), stored(t, "band.css", bandSource))
	if !errors.Is(err, errOutside) {
		t.Fatalf("a source out of range gave %v", err)
	}
	if !strings.Contains(printed, "dim: ink on the ground: too dim to read at length") {
		t.Errorf("the failure is not named:\n%s", printed)
	}
}

// --policy swaps the range: the same source passes one and fails the next.
func TestACheckedSourceIsHeldToTheRangeGiven(t *testing.T) {
	source := stored(t, "band.css", bandSource)
	printed, err := checked(t, "--policy", stored(t, "newer.json", newerRange), source)
	if !errors.Is(err, errOutside) || !strings.Contains(printed, "glare") {
		t.Errorf("the newer range gave %v, %q", err, printed)
	}
}

// The anti-tests: check with nothing to check, a flag with no file, a range
// that is not there or not a range, a source that is not there, and a
// source that cannot make a scheme are errors, not a pass.
func TestCheckRefusesWhatItCannotHold(t *testing.T) {
	source := stored(t, "band.css", bandSource)
	thin := stored(t, "thin.css", ":root {\n  --ink: #ffffff;\n}\n")
	for name, arguments := range map[string][]string{
		"nothing":           {},
		"flag with nothing": {"--policy"},
		"flag and no file":  {"--policy", source},
		"flag not known":    {"--wider", source},
		"range not there":   {"--policy", "/no/such/range.json", source},
		"range not a range": {"--policy", stored(t, "bad.json", "{"), source},
		"source not there":  {"/no/such/source.css"},
		"source too thin":   {thin},
	} {
		if _, err := checked(t, arguments...); err == nil || errors.Is(err, errOutside) {
			t.Errorf("%s was not an error", name)
		}
	}
}
