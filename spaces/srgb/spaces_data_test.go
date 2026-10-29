package srgb_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// wrapCase is a hue and the hue it is the same as.
type wrapCase struct {
	HSV    srgb.HSV
	SameAs srgb.HSV `json:"same as"`
}

// picked is the numbers these tests pick, from testdata/spaces.json.
type picked struct {
	Gamut swatch.Gamut
	Curve struct {
		Break       float64
		LinearBreak float64 `json:"linear break"`
		Slope       float64
		Offset      float64
		Scale       float64
		Power       float64
		Tolerance   float64
	}
	Slack struct {
		Below, Above float64
		Reach        int     `json:"search reach"`
		WayOver      float64 `json:"way over"`
	}
	Wrap           []wrapCase
	WheelTolerance float64 `json:"wheel tolerance"`
	Dark           struct {
		AtThreshold   srgb.RGB `json:"at the threshold"`
		PastThreshold srgb.RGB `json:"past the threshold"`
		Green         srgb.RGB `json:"green at the threshold"`
		GreenHue      float64  `json:"green hue"`
	}
	RGBA struct {
		Levels srgb.RGB
		Want   struct {
			Red   uint32 `json:"R"`
			Green uint32 `json:"G"`
			Blue  uint32 `json:"B"`
			Alpha uint32 `json:"A"`
		}
	}
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
