package main

import (
	"encoding/json"
	"os"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/spaces/ok"
)

// picked is the numbers these tests pick, from testdata/hue.json: more
// chroma than any colour the lamps make has, and the part of the eye's
// distance the faint chromas start from, so the band where hue is dropped
// is drawn from often.
type picked struct {
	MostChroma float64 `json:"most chroma"`
	FaintFrom  float64 `json:"faint from"`
}

// pickedNumbers reads the test numbers, or the test stops.
func pickedNumbers(t interface{ Fatal(...any) }) picked {
	raw, err := os.ReadFile("testdata/hue.json")
	if err != nil {
		t.Fatal(err)
	}
	var numbers picked
	if err := json.Unmarshal(raw, &numbers); err != nil {
		t.Fatal(err)
	}
	return numbers
}

// fullCircle is the degrees round the hue wheel. It is defined by the
// circle.
const fullCircle = 360

// The property: a colour written as css and read back is the same colour to
// an eye, at any lightness, chroma and hue, including the low chroma where
// the hue is dropped to none. Dropping a hue moves a colour by up to twice
// its chroma, so none may only stand for chroma well below the eye's
// distance.
func TestAColourSurvivesBeingWrittenAsCSS(t *testing.T) {
	rapid.Check(t, survivesCSS)
}

// survivesCSS is one draw of the property above.
func survivesCSS(t *rapid.T) {
	most := pickedNumbers(t).MostChroma
	survives(t, rapid.Float64Range(ok.Black.L, most))
}

// The same, for the faint chromas where the hue is dropped: from a part of
// the eye's distance to the eye's distance.
func TestAFaintColourSurvivesBeingWrittenAsCSS(t *testing.T) {
	rapid.Check(t, survivesFaintCSS)
}

// survivesFaintCSS is one draw of the property above.
func survivesFaintCSS(t *rapid.T) {
	from := ok.Eye / pickedNumbers(t).FaintFrom
	survives(t, rapid.Float64Range(from, ok.Eye))
}

// survives checks one colour, with its chroma drawn from the source given.
func survives(t *rapid.T, chromas *rapid.Generator[float64]) {
	var original ok.OKLCH
	original.L = rapid.Float64Range(ok.Black.L, ok.White.L).Draw(t, "lightness")
	original.C = chromas.Draw(t, "chroma")
	original.H = rapid.Float64Range(ok.Black.L, fullCircle).Draw(t, "hue")
	written := original.String()
	back, err := parse(written)
	if err != nil {
		t.Fatalf("%s did not parse: %v", written, err)
	}
	if moved := ok.Distance(original.Rect(), ok.FromSwatch(back)); moved >= ok.Eye {
		t.Fatalf("%s moved the colour by %.4f, the eye sees %.4f",
			written, moved, ok.Eye)
	}
}
