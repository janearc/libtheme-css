package css_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
)

// These tests hold Named. It lists custom properties that are not hex
// colours and that the caller reads. Each has its line and its value as
// written, so a role in another form of colour is refused and not lost.

// everyName says that the caller reads every name.
func everyName(string) bool { return true }

// firstLine is the line of the first property in the sources below, under
// the :root { line. It is defined by those sources.
const firstLine = 2

// entry is a property of a source: its name and its value as written.
type entry struct{ name, value string }

// sourceOf is a :root block with a property on each line. It also gives what
// Named is to say of the ones that are not hex colours.
func sourceOf(entries []entry) (string, []string) {
	var source strings.Builder
	source.WriteString(":root {\n")
	var want []string
	for place, each := range entries {
		fmt.Fprintf(&source, "  --%s: %s;\n", each.name, each.value)
		if !strings.HasPrefix(each.value, "#") {
			want = append(want, fmt.Sprintf(
				"line %d: --%s: %s is not a #rgb or #rrggbb",
				place+firstLine, each.name, each.value))
		}
	}
	source.WriteString("}\n")
	return source.String(), want
}

// Other forms of colour are named. So are a font and an empty value. They
// come in file order, each with the line its name is on and its value as
// written. The hex colours between them are not named.
func TestNamedListsWhatIsNotAHexColour(t *testing.T) {
	source, want := sourceOf([]entry{
		{"ink", "oklch(70% 0.1 200)"},
		{"a", "#fff"},
		{"font", "Helvetica"},
		{"empty", ""},
		{"x", "rgb(1 2 3) /* c */"},
		{"b", "#112233"},
	})
	if got := css.Read(source).Named(everyName); !slices.Equal(got, want) {
		t.Errorf("named\n%q\nwant\n%q", got, want)
	}
}

// Only the names the caller reads are listed. The anti-test is the caller
// that reads none.
func TestNamedListsOnlyTheNamesTheCallerReads(t *testing.T) {
	source, _ := sourceOf([]entry{{"ink", "red"}, {"font", "Helvetica"}})
	got := css.Read(source).Named(onlyInk)
	if want := "line 2: --ink: red is not a #rgb or #rrggbb"; !slices.Equal(got, []string{want}) {
		t.Errorf("named %q, want %q", got, want)
	}
	if got := css.Read(source).Named(none); len(got) != nothing {
		t.Errorf("a caller that reads nothing was told %q", got)
	}
}

// onlyInk says that the caller reads the name ink and no other.
func onlyInk(name string) bool { return name == "ink" }

// none says that the caller reads no name.
func none(string) bool { return false }

// Properties after the start of the file are found without a space before
// them. So are those after a semicolon or a brace. Colons may have space
// before them.
func TestNamedFindsAPropertyAfterAnyBoundary(t *testing.T) {
	source := "--first: red;--second :blue;a{--third:green}"
	got := css.Read(source).Named(everyName)
	want := []string{
		"line 1: --first: red is not a #rgb or #rrggbb",
		"line 1: --second: blue is not a #rgb or #rrggbb",
		"line 1: --third: green is not a #rgb or #rrggbb",
	}
	if !slices.Equal(got, want) {
		t.Errorf("named %q, want %q", got, want)
	}
}

// The anti-test: a double hyphen inside a word, a name in a comment and a
// hex colour are not other properties.
func TestNamedLeavesOutWhatIsNotAProperty(t *testing.T) {
	source := ".btn--ink:hover { color: red }\n/* --dim: red; */\n" +
		":root { --ink: #fff; --dim:#000 }\n"
	if got := css.Read(source).Named(everyName); len(got) != nothing {
		t.Errorf("named %q", got)
	}
}
