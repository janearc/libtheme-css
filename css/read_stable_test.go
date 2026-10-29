package css_test

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/css"
)

// stableNumbers are the numbers these tests pick, from testdata/stable.json.
type stableNumbers struct {
	CycleLength  int   `json:"cycle length"`
	Cycle        []int `json:"cycle"`
	LeastMembers int   `json:"least members"`
	MostMembers  int   `json:"most members"`
	LeastNumber  int   `json:"least number"`
	MostNumber   int   `json:"most number"`
	GreyStep     int   `json:"grey step"`
	TextOrder    []int `json:"text order"`
}

// stablePicked reads the numbers, or the test stops.
func stablePicked(t testing.TB) stableNumbers {
	t.Helper()
	var numbers stableNumbers
	raw, err := os.ReadFile("testdata/stable.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &numbers); err != nil {
		t.Fatal(err)
	}
	return numbers
}

// everyMatch asks regexp for every match, not a count of them. It is
// defined by the regexp package: a negative count is all.
const everyMatch = -1

// hexRe finds the colours of a gradient, in order.
var hexRe = regexp.MustCompile(`#[0-9a-f]{6}`)

// stopsOf is the colours of the ramp called "m" in the order its stops run,
// or nothing when there is no such ramp.
func stopsOf(read css.Colourway) []string {
	ramp, has := read.Ramps["m"]
	if !has {
		return nil
	}
	return hexRe.FindAllString(ramp.String(hexOf), everyMatch)
}

// stopsRead is the stops of the ramp a source makes.
func stopsRead(source string) []string {
	return stopsOf(css.Read(source))
}

// familyOf is a source whose member i has the number numbers[i]. Each name
// has leading zeros so that it is its own. Each member has the grey of its
// place in the file. It also gives the greys in the order a stable sort by
// number puts them.
func familyOf(numbers []int, step int) (string, []string) {
	var source strings.Builder
	source.WriteString(":root {\n")
	greys := make([]string, len(numbers))
	for place, number := range numbers {
		grey := place * step
		greys[place] = fmt.Sprintf("#%02x%02x%02x", grey, grey, grey)
		fmt.Fprintf(&source, "  --m-%s%d: %s;\n",
			strings.Repeat("0", place), number, greys[place])
	}
	source.WriteString("}\n")
	var order []string
	distinct := slices.Sorted(slices.Values(numbers))
	for _, number := range slices.Compact(distinct) {
		for place, each := range numbers {
			if each == number {
				order = append(order, greys[place])
			}
		}
	}
	return source.String(), order
}

// Members that share a number keep the file's order in a long family. It is
// past the size where a sort by insertion hides the fault. The numbers
// cycle through 3 then 1 then 2.
func TestMembersOfOneNumberKeepTheFilesOrderInALongFamily(t *testing.T) {
	picked := stablePicked(t)
	numbers := make([]int, picked.CycleLength)
	for place := range numbers {
		numbers[place] = picked.Cycle[place%len(picked.Cycle)]
	}
	source, want := familyOf(numbers, picked.GreyStep)
	if got := stopsRead(source); !slices.Equal(got, want) {
		t.Errorf("stops\n%v\nwant\n%v", got, want)
	}
}

// stability is the numbers the property below is drawn from.
type stability struct{ picked stableNumbers }

// keepsTheFilesOrder is the property. Numbers are drawn shuffled. The stops
// of one number are in the file's order.
func (drawn stability) keepsTheFilesOrder(t *rapid.T) {
	picked := drawn.picked
	count := rapid.IntRange(picked.LeastMembers, picked.MostMembers).Draw(t, "members")
	number := rapid.IntRange(picked.LeastNumber, picked.MostNumber)
	numbers := rapid.SliceOfN(number, count, count).Draw(t, "numbers")
	source, want := familyOf(numbers, picked.GreyStep)
	if got := stopsRead(source); !slices.Equal(got, want) {
		t.Fatalf("numbers %v\nstops\n%v\nwant\n%v", numbers, got, want)
	}
}

// The same holds for any family past twelve members, whatever the numbers.
func TestMembersOfOneNumberKeepTheFilesOrderWhateverTheNumbers(t *testing.T) {
	rapid.Check(t, stability{stablePicked(t)}.keepsTheFilesOrder)
}

// The numbers in the names are read as numbers. They run in numeric order,
// where text would put 10 and 100 before 8. This is the anti-test to
// the above, which takes the file's order as the order of the names.
func TestNumbersAreReadAsNumbersNotText(t *testing.T) {
	picked := stablePicked(t)
	source, want := familyOf(picked.TextOrder, picked.GreyStep)
	if got := stopsRead(source); !slices.Equal(got, want) {
		t.Errorf("stops %v, want %v", got, want)
	}
}
