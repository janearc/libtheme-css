package radiation_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/janearc/libtheme-css/spaces/radiation"
)

// picked is the numbers these tests pick, from testdata/spaces.json.
type picked struct {
	Named struct {
		Gamma       radiation.Range `json:"gamma ray"`
		XRay        radiation.Range `json:"x-ray"`
		ExtremeUV   radiation.Range `json:"extreme ultraviolet"`
		Ultraviolet radiation.Range `json:"ultraviolet"`
		Infrared    radiation.Range
		Microwave   radiation.Range
		Radio       radiation.Range
	} `json:"named ranges"`
	Fermi struct {
		Highest   float64 `json:"highest ev"`
		Lowest    float64 `json:"lowest ev"`
		Tolerance float64
	}
	Ladders struct{ Energy, Frequency []float64 }
	Profile struct {
		Flat     radiation.Profile
		BackRend radiation.Profile `json:"backwards render"`
		BackSrc  radiation.Profile `json:"backwards source"`
		Sound    radiation.Profile
		Short    radiation.Profile
		Outside  radiation.Profile
		InsideNM float64 `json:"inside nm"`
		BelowNM  float64 `json:"below nm"`
		AboveNM  float64 `json:"above nm"`
	} `json:"profiles"`
	Of struct {
		Wavelengths []int
		DeepViolet  int `json:"deep violet nm"`
		Between     struct {
			Low  int `json:"low nm"`
			High int `json:"high nm"`
		}
		Within float64                           `json:"within nm"`
		Half   float64                           `json:"half nm"`
		Off    struct{ Degrees, Radius float64 } `json:"off the sweep"`
		OneRow int                               `json:"one row"`
		FarRed struct {
			Short  int     `json:"short nm"`
			Long   int     `json:"long nm"`
			Within float64 `json:"within nm"`
		} `json:"far red"`
		Magenta  string
		HalfTurn float64 `json:"half turn degrees"`
	}
	Light struct {
		LineNM    float64 `json:"line nm"`
		Power     float64
		Lead      float64         `json:"lead nm"`
		Squeezed  radiation.Range `json:"squeezed render"`
		CellNM    int             `json:"cell nm"`
		Tolerance float64
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
