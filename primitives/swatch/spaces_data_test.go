package swatch_test

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
)

// row is one row of the observer's table as testdata/spaces.json holds it.
type row struct {
	NM   int
	Xbar float64 `json:"x"`
	Ybar float64 `json:"y"`
	Zbar float64 `json:"z"`
}

// hairTriangle is a triangle and two points near one of its edges.
type hairTriangle struct {
	Turn string
	swatch.Gamut
	HairOut     swatch.XY `json:"a hair out"`
	TwoHairsOut swatch.XY `json:"two hairs out"`
}

// picked is the numbers these tests pick, from testdata/spaces.json.
type picked struct {
	Rows struct {
		Tolerance float64
		At        []row
	}
	Planck struct {
		Planck    float64 `json:"h"`
		Light     float64 `json:"c"`
		Boltzmann float64 `json:"k"`
		Kelvin    float64
		NM        int
		Tolerance float64
	}
	Wien struct {
		Constant float64
		Kelvin   float64
		SlackNM  float64 `json:"slack nm"`
	}
	Kelvins struct {
		One       float64
		Zero      float64
		BelowZero float64 `json:"below zero"`
	}
	D65 struct {
		XY           swatch.XY
		Luminance    float64
		XAtOne       float64 `json:"x at one"`
		ZAtOne       float64 `json:"z at one"`
		ToleranceOne float64 `json:"tolerance at one"`
		Tolerance    float64
	} `json:"d65 chromaticity"`
	Dim struct {
		XYZ struct {
			CieX float64 `json:"x"`
			CieY float64 `json:"y"`
			CieZ float64 `json:"z"`
		}
		Chroma    swatch.XY `json:"chromaticity"`
		Tolerance float64
	} `json:"dim swatch"`
	Gamut    swatch.Gamut
	UnitEdge swatch.Gamut `json:"unit edge gamut"`
	Hair     struct {
		Triangles []hairTriangle
	}
	Grid struct{ From, To, Step, Margin float64 }
	Foot struct {
		Along     float64
		OutBy     float64 `json:"out by"`
		Tolerance float64
	}
	Past struct{ East, South float64 } `json:"past the red corner"`
	Blue struct{ Points []swatch.XY }  `json:"past the blue corner"`
}

// pickedNumbers reads the test numbers, or the test stops.
func pickedNumbers(t *testing.T) picked {
	t.Helper()
	raw, err := os.ReadFile("testdata/spaces.json")
	if err != nil {
		t.Fatal(err)
	}
	var numbers picked
	if err := json.Unmarshal(raw, &numbers); err != nil {
		t.Fatal(err)
	}
	return numbers
}

// Numbers the swatch tests use that the arithmetic or the SI defines.
const (
	// none is nothing. It is defined by the scale.
	none = 0
	// whole is all of a scale. It is defined by the scale.
	whole = 1
	// twice is the factor in doubling. It is defined by the arithmetic.
	twice = 2
	// fifth is the power of wavelength in Planck's law. It is defined by Planck's law.
	fifth = 5
	// nanometre is metres in a nanometre. It is defined by the SI prefix.
	nanometre = 1e-9
)

// near says whether two numbers agree to a relative tolerance.
func near(got, want, tolerance float64) bool {
	gap := math.Abs(got - want)
	return gap <= tolerance*math.Abs(want)
}
