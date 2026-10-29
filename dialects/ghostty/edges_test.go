package ghostty_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// edges is testdata/edges.json: where palette-N entries start and end.
type edges struct {
	First int `json:"first named"`
	Last  int `json:"last named"`
	Step  int
}

// edgeNumbers reads testdata/edges.json, or the test stops.
func edgeNumbers(t *testing.T) edges {
	t.Helper()
	raw, err := os.ReadFile("testdata/edges.json")
	if err != nil {
		t.Fatal(err)
	}
	var numbers edges
	if err := json.Unmarshal(raw, &numbers); err != nil {
		t.Fatal(err)
	}
	return numbers
}

// ink is the colour the tests set, and look for.
func ink() swatch.Swatch { return srgb.MustHex("#010203").Swatch() }

// completeRoles is the roles Of cannot do without, and one more.
func completeRoles(extra string) *css.Sheet {
	roles := css.New()
	roles.Set("ground", ink())
	roles.Set("ink", ink())
	for _, name := range ghostty.Ansi {
		roles.Set(name, ink())
	}
	roles.Set(extra, ink())
	return roles
}

// paletteLine is the line a theme writes for one entry.
func paletteLine(number int) string {
	return fmt.Sprintf("palette = %d=#010203\n", number)
}

// themeOf is a theme holding the settings and entries given.
func themeOf(entries ...int) ghostty.Theme {
	var theme ghostty.Theme
	theme.Settings = map[string]swatch.Swatch{"background": ink()}
	theme.Palette = map[int]swatch.Swatch{}
	for _, number := range entries {
		theme.Palette[number] = ink()
	}
	return theme
}

// Of takes palette-N from the first named entry to the last, and
// refuses the number one step outside each end.
func TestOfTakesTheFirstToTheLastExactly(t *testing.T) {
	numbers := edgeNumbers(t)
	inside := []int{numbers.First, numbers.First + numbers.Step,
		numbers.Last - numbers.Step, numbers.Last}
	outside := []int{numbers.First - numbers.Step, numbers.Last + numbers.Step}
	for _, number := range inside {
		roles := completeRoles(fmt.Sprintf("palette-%d", number))
		theme, err := ghostty.Of(roles)
		if _, found := theme.Palette[number]; err != nil || !found {
			t.Errorf("palette-%d: %v, kept %v", number, err, found)
		}
	}
	for _, number := range outside {
		roles := completeRoles(fmt.Sprintf("palette-%d", number))
		if _, err := ghostty.Of(roles); err == nil {
			t.Errorf("palette-%d was taken", number)
		}
	}
}

// Roles names the entries from the first named entry up as palette-N.
// The entry under it is one of the sixteen, named as itself.
func TestRolesStartPaletteNamesAtTheFirstNamedEntry(t *testing.T) {
	numbers := edgeNumbers(t)
	under := numbers.First - numbers.Step
	theme := themeOf(under, numbers.First, numbers.First+numbers.Step)
	want := []string{"ground", ghostty.Ansi[under],
		fmt.Sprintf("palette-%d", numbers.First),
		fmt.Sprintf("palette-%d", numbers.First+numbers.Step)}
	if got := theme.Roles().Names(); !slices.Equal(got, want) {
		t.Errorf("roles %q, want %q", got, want)
	}
}

// Read skips a line with no equals sign, even one that is a key's name.
// The bare word "palette" is not a setting.
func TestReadSkipsALineWithNoEquals(t *testing.T) {
	text := "palette\nbackground\nforeground = #0a0b0c\nnothing here\n"
	theme, err := ghostty.Read(strings.NewReader(text))
	if err != nil {
		t.Fatalf("a bare word was read as a setting: %v", err)
	}
	if got := slices.Collect(maps.Keys(theme.Settings)); !slices.Equal(got, []string{"foreground"}) {
		t.Errorf("settings read: %q", got)
	}
	if got := slices.Collect(maps.Keys(theme.Palette)); got != nil {
		t.Errorf("entries read: %v", got)
	}
}

// Read takes the last palette entry and refuses the one after it.
func TestReadTakesTheLastEntryAndNoMore(t *testing.T) {
	numbers := edgeNumbers(t)
	last := paletteLine(numbers.Last)
	theme, err := ghostty.Read(strings.NewReader(last))
	if _, found := theme.Palette[numbers.Last]; err != nil || !found {
		t.Errorf("entry %d: %v, kept %v", numbers.Last, err, found)
	}
	past := paletteLine(numbers.Last + numbers.Step)
	if _, err := ghostty.Read(strings.NewReader(past)); err == nil {
		t.Errorf("entry %d was taken", numbers.Last+numbers.Step)
	}
}

// Write puts each entry once. It leaves a blank line before the entries
// over the sixteen only when there are some, even just one.
func TestWriteSeparatesTheHighEntriesOnlyWhenThereAreSome(t *testing.T) {
	numbers := edgeNumbers(t)
	under, first := numbers.First-numbers.Step, numbers.First
	next := numbers.First + numbers.Step
	head := "# h\n\nbackground = #010203\n\n" + paletteLine(under)
	for name, one := range map[string]struct {
		entries []int
		want    string
	}{
		"none above": {[]int{under}, head},
		"one above":  {[]int{under, first}, head + "\n" + paletteLine(first)},
		"two above": {[]int{under, first, next},
			head + "\n" + paletteLine(first) + paletteLine(next)},
	} {
		var out strings.Builder
		if err := themeOf(one.entries...).Write(&out, "h"); err != nil {
			t.Fatal(err)
		}
		if out.String() != one.want {
			t.Errorf("%s:\n%q\nwant\n%q", name, out.String(), one.want)
		}
	}
}
