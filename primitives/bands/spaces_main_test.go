package bands_test

import (
	"encoding/json"
	"os"
	"testing"
)

// picked is the numbers these tests pick, from testdata/spaces.json.
type picked struct {
	CoolLamp   float64 `json:"cool lamp"`
	WarmLamp   float64 `json:"warm lamp"`
	MiddleLamp float64 `json:"middle lamp"`
	Shift      int     `json:"shift"`
	Visible    struct {
		Short float64 `json:"short nm"`
		Long  float64 `json:"long nm"`
		First int     `json:"first band"`
		Last  int     `json:"last band"`
		Width int     `json:"width"`
	}
	WeightBand int     `json:"weight band"`
	LinePower  float64 `json:"line power"`
	LineNM     float64 `json:"line nm"`
	Tolerance  float64 `json:"relative tolerance"`
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
