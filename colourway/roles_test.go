package colourway_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/colourway"
)

// readText reads a source file with this text, under the name dusk.
func readText(t *testing.T, text string) (colourway.Source, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dusk.css")
	if err := os.WriteFile(path, []byte(text), private); err != nil {
		t.Fatal(err)
	}
	return colourway.Read(path)
}

// Roles written in another form of colour are refused. The complaint names
// the role and the line and the value. Without it the role would be missing
// from every file rendered and nobody told. The roles are of each kind a
// dialect reads. They are general and terminal and ghostty and claude code
// and neovim and numbered and overridden and markdown roles.
func TestARoleInAnotherFormOfColourIsRefused(t *testing.T) {
	for name, value := range map[string]string{
		"ink":          "oklch(70% 0.1 200)",
		"ground":       "rgb(1 2 3)",
		"red":          "hsl(0 100% 50%)",
		"bright-black": "gray",
		"cursor":       "var(--ink)",
		"claude-text":  "oklch(90% 0 0)",
		"cursor-line":  "rgb(0 0 0 / 50%)",
		"palette-252":  "rgb(1 2 3)",
		"nvim-ink":     "red",
		"md-body":      "hsl(10 10% 10%)",
		"md-link-text": "red",
		"err":          "",
	} {
		text := ":root {\n  --font: Helvetica;\n  --" + name + ": " + value + ";\n}\n"
		_, err := readText(t, text)
		want := "line 3: --" + name + ": " + value + " is not a #rgb or #rrggbb"
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("--%s: %s gave %v, want %q", name, value, err, want)
		}
	}
}

// The anti-test: properties that are not roles are left alone in any form,
// and the same roles in hex are read. Only the names decide. A name of
// digits alone is no palette number: only palette- and digits is one.
func TestAPropertyThatIsNoRoleIsLeftAlone(t *testing.T) {
	text := ":root {\n  --font: Helvetica;\n  --ink-shadow: 0 1px black;\n" +
		"  --nvim-nonsense: red;\n  --palette-x: red;\n  --42: red;\n" +
		"  --inky: rgb(1 2 3);\n" +
		"  --ink: #112233;\n  --palette-252: #445566;\n}\n"
	source, err := readText(t, text)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := source.Roles.Names(), []string{"ink", "palette-252"}; !slices.Equal(got, want) {
		t.Errorf("roles read %v, want %v", got, want)
	}
}
