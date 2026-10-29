package readable_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/readable"
)

// rangeNumbers are the five numbers of a range file.
type rangeNumbers struct {
	Low      float64
	High     float64
	StateMin float64 `json:"state_min"`
	Gap      float64
	Neutral  float64
}

// rangeCase is a range file and whether it should be read.
type rangeCase struct {
	rangeNumbers
	Name string
	Read bool
}

// huePair is a ground and an ink, by hex.
type huePair struct {
	Name, Ground, Ink string
}

// edgeFile is testdata/edges.json.
type edgeFile struct {
	Typical     rangeNumbers
	Room        float64
	Unreachable rangeNumbers
	RangeCases  []rangeCase       `json:"range cases"`
	HuePairs    []huePair         `json:"hue pairs"`
	EdgeGround  string            `json:"edge ground"`
	Light       map[string]string `json:"light scheme"`
}

// edgeData reads testdata/edges.json, or the test stops.
func edgeData(t *testing.T) edgeFile {
	t.Helper()
	raw, err := os.ReadFile("testdata/edges.json")
	if err != nil {
		t.Fatal(err)
	}
	var data edgeFile
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	return data
}

// text is the range file of the numbers.
func (numbers rangeNumbers) text() string {
	const form = `{"contrast": {"low": %g, "high": %g, "state_min": %g}, ` +
		`"hue": {"min_separation_degrees": %g, "neutral_saturation": %g}}`
	return fmt.Sprintf(form, numbers.Low, numbers.High, numbers.StateMin,
		numbers.Gap, numbers.Neutral)
}

// policyOf is the policy of the numbers, or the test stops.
func policyOf(t *testing.T, numbers rangeNumbers) readable.Policy {
	t.Helper()
	policy, err := readText(numbers.text())
	if err != nil {
		t.Fatalf("%+v: %v", numbers, err)
	}
	return policy
}

// hasLine is whether the report has a line with the text in it.
func hasLine(report readable.Report, text string) bool {
	return strings.Contains(joined(report), text)
}

// ratio is the contrast of one colour of a scheme on another.
func ratio(scheme nvim.Scheme, ink, ground string) float64 {
	return nvim.Contrast(scheme.Colours[ink], scheme.Colours[ground])
}

// Read takes the ranges at the edge of what nests, and refuses those one
// step over it.
func TestARangeAtItsEdgesIsReadAndOneStepOverIsNot(t *testing.T) {
	for _, one := range edgeData(t).RangeCases {
		_, err := readText(one.text())
		if (err == nil) != one.Read {
			t.Errorf("%s: error %v, read should be %v", one.Name, err, one.Read)
		}
	}
}

// Read refuses a part of the file that is a number where an object
// belongs. It is not read as if it held the number.
func TestANumberWhereAnObjectShouldBeIsRefused(t *testing.T) {
	typical := edgeData(t).Typical
	contrast := fmt.Sprintf(`{"contrast": %g, "hue": {"min_separation_degrees": %g, "neutral_saturation": %g}}`,
		typical.Neutral, typical.Gap, typical.Neutral)
	hue := fmt.Sprintf(`{"contrast": {"low": %g, "high": %g, "state_min": %g}, "hue": %g}`,
		typical.Low, typical.High, typical.StateMin, typical.Neutral)
	for name, text := range map[string]string{"contrast": contrast, "hue": hue} {
		if policy, err := readText(text); err == nil {
			t.Errorf("%s a number was read as %v", name, policy)
		}
	}
}

// Read names what is wrong with a range file, and names the first thing.
func TestARangeNamesTheFirstThingWrong(t *testing.T) {
	_, err := readText("{")
	if err == nil || !strings.Contains(err.Error(), "readable: the range:") {
		t.Errorf("not json gave %v", err)
	}
	_, err = readText(`{"contrast": {"high": 12.5}, "hue": {}}`)
	if err == nil || err.Error() != "readable: no number at contrast.low" {
		t.Errorf("two numbers missing gave %v", err)
	}
}
