package claude_test

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/claude"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// slotRow is one slot as testdata/slots.json holds it.
type slotRow struct {
	Role   string
	Number int
	Use    string
	From   string
	Toward float64
}

// slotFile is testdata/slots.json: the table, and the numbers around
// the edge where claude code's slots start to be named by role.
type slotFile struct {
	Slots []slotRow
	Edges struct{ First, Step int }
}

// slotData reads testdata/slots.json, or the test stops.
func slotData(t *testing.T) slotFile {
	t.Helper()
	raw, err := os.ReadFile("testdata/slots.json")
	if err != nil {
		t.Fatal(err)
	}
	var data slotFile
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	return data
}

// held is a value to put back where it was taken from.
type held[Kind any] struct {
	target *Kind
	value  Kind
}

// hold takes the value at target, for restore to put back.
func hold[Kind any](target *Kind) held[Kind] { return held[Kind]{target, *target} }

// restore puts the held value back.
func (taken held[Kind]) restore() { *taken.target = taken.value }

// aroundTheEdge narrows the slots to three. They are numbered one step
// under the first named number, at it and one step over it. The test gets
// those numbers and the three roles. The slots are put back at the end.
func aroundTheEdge(t *testing.T) (numbers []int, roles []string) {
	t.Helper()
	edges := slotData(t).Edges
	t.Cleanup(hold(&claude.Slots).restore)
	claude.Slots = slices.Clone(claude.Slots[:edges.Step+edges.Step+edges.Step])
	numbers = []int{edges.First - edges.Step, edges.First, edges.First + edges.Step}
	for place, number := range numbers {
		claude.Slots[place].Number = number
		roles = append(roles, claude.Slots[place].Role)
	}
	return numbers, roles
}

// colourOf is a colour for the tests to set, and to look for.
func colourOf() swatch.Swatch { return srgb.MustHex("#123456").Swatch() }

// Slots is claude code's own table. Each entry is a fact about what it
// draws, so each is pinned.
func TestSlotsAreClaudeCodesTable(t *testing.T) {
	want := slotData(t).Slots
	var got []slotRow
	for _, slot := range claude.Slots {
		got = append(got, slotRow(slot))
	}
	if !slices.Equal(got, want) {
		t.Errorf("the table is\n%v\nwant\n%v", got, want)
	}
}

// Into skips the slots below the first one named by role, which the
// terminal already names. The first and those over it are said as
// palette-N.
func TestIntoSkipsBelowTheFirstNamedOnly(t *testing.T) {
	numbers, roles := aroundTheEdge(t)
	first := slotData(t).Edges.First
	sheet := css.New()
	for _, role := range roles {
		sheet.Set(role, colourOf())
	}
	said, err := claude.Into(sheet)
	if err != nil {
		t.Fatal(err)
	}
	for _, number := range numbers {
		name := fmt.Sprintf("palette-%d", number)
		if _, found := said.Get(name); found != (number >= first) {
			t.Errorf("%s said: %v", name, found)
		}
	}
}

// Of names a palette's colours back from the first named number up.
// The number under it stays the terminal's.
func TestOfNamesFromTheFirstNamedUp(t *testing.T) {
	numbers, roles := aroundTheEdge(t)
	first := slotData(t).Edges.First
	palette := map[int]swatch.Swatch{}
	for _, number := range numbers {
		palette[number] = colourOf()
	}
	named := claude.Of(palette)
	for place, role := range roles {
		if _, found := named.Get(role); found != (numbers[place] >= first) {
			t.Errorf("%s named: %v", role, found)
		}
	}
}
