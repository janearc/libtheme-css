package ok_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/janearc/libtheme-css/spaces/ok"
)

// wheelCase is a colour and the hue it is at.
type wheelCase struct {
	Lab ok.OKLab
	Hue float64
}

// picked is the numbers these tests pick, from testdata/spaces.json.
type picked struct {
	Fit struct {
		Colour    ok.OKLCH
		Halvings  int
		BoxEdge   float64 `json:"box edge"`
		Tolerance float64
	}
	Polar struct {
		Grey         ok.OKLCH
		AHair        ok.OKLab `json:"a hair"`
		ExactlyExact ok.OKLab `json:"exactly exact"`
		Wheel        []wheelCase
		DegreeSlack  float64 `json:"tolerance degrees"`
	}
	White struct{ Tolerance float64 }
	Same  struct{ Apart ok.OKLab }
	Reach struct {
		Cols, Rows int
		Tolerance  float64
		BoxEdge    float64 `json:"box edge"`
		Defaults   struct {
			HueFrom   float64 `json:"hue from"`
			HueTo     float64 `json:"hue to"`
			LightFrom float64 `json:"light from"`
			LightTo   float64 `json:"light to"`
			Chroma    float64
		}
		Probe  float64 `json:"probe chroma"`
		Half   float64 `json:"half chroma"`
		Small  struct{ Two, One int }
		Hues   struct{ From, To float64 }
		Far    struct{ Light, Dark ok.OKLab }
		Twin   struct{ Col, Row int }
		OffTop struct{ Col, Row int } `json:"off the grid"`
	}
}

// readPicked reads the test numbers.
func readPicked() (picked, error) {
	var numbers picked
	raw, err := os.ReadFile("testdata/spaces.json")
	if err != nil {
		return numbers, err
	}
	return numbers, json.Unmarshal(raw, &numbers)
}

// pickedNumbers reads the test numbers, or the test stops.
func pickedNumbers(t *testing.T) picked {
	t.Helper()
	numbers, err := readPicked()
	if err != nil {
		t.Fatal(err)
	}
	return numbers
}
