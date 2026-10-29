package nvim_test

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/janearc/libtheme-css/dialects/nvim"
)

// The counts the edge tests below use. Each is defined by the file
// layout or by the sixteen.
const (
	// intoTable is how many lines past the table's own text the first
	// row of the terminal table sits: the opener, then the row. It is
	// defined by the layout of termOf.
	intoTable = 2
	// rowsOfEight is the rows the sixteen are written in. It is defined
	// by the sixteen over eight.
	rowsOfEight = 2
)

// tableOf is a colour table with the entries given.
func tableOf(entries ...string) string {
	return "local c = {\n" + strings.Join(entries, "\n") + "\n}\n"
}

// readTable reads a colour table with the entries given.
func readTable(entries ...string) (nvim.Scheme, error) {
	text := tableOf(entries...)
	return nvim.Read(strings.NewReader(text))
}

// Read calls a scheme light only when its ground is lighter than its
// ink. Equal is not lighter.
func TestEqualGroundAndInkIsNotLight(t *testing.T) {
	same, err := readTable(`  bg = "#808080",`, `  ink = "#808080",`)
	if err != nil || same.Light {
		t.Errorf("equal ground and ink: %v, light %v", err, same.Light)
	}
	light, err := readTable(`  bg = "#ffffff",`, `  ink = "#808080",`)
	if err != nil || !light.Light {
		t.Errorf("a white ground over grey ink: %v, light %v", err, light.Light)
	}
}

// The anti-test: with only one of the ground and the ink the scheme is
// not light, whichever colour is the one given.
func TestLightNeedsBothGroundAndInk(t *testing.T) {
	for name, table := range map[string]string{
		"ground only": tableOf(`  bg = "#ffffff",`),
		"ink only":    tableOf(`  ink = "#000000",`)} {
		scheme, err := nvim.Read(strings.NewReader(table))
		if err != nil || scheme.Light {
			t.Errorf("%s: %v, light %v", name, err, scheme.Light)
		}
	}
}

// Read refuses a file whose read fails after giving some of it. The
// failure comes back, and the part read is not used.
func TestAFileThatFailsPartwayIsRefused(t *testing.T) {
	broken := errors.New("the disk went")
	given := strings.NewReader(tableOf(`  bg = "#ffffff",`, `  ink = "#000000",`))
	failing := iotest.ErrReader(broken)
	_, err := nvim.Read(io.MultiReader(given, failing))
	if !errors.Is(err, broken) {
		t.Errorf("a half-read file gave %v", err)
	}
}

// Read names a bad line by its own number, counting from one.
func TestTheBadLineIsNumberedFromOne(t *testing.T) {
	line := strings.Count(tabled, "\n") + intoTable
	_, err := readTerm(tooMany)
	if want := fmt.Sprintf("nvim: line %d: more than sixteen", line); err == nil ||
		!strings.Contains(err.Error(), want) {
		t.Errorf("the error is %v, want %q", err, want)
	}
}

// Write puts the sixteen in two rows of exactly eight quoted colours.
// Nothing follows the last colour of a row but its comma.
func TestTheSixteenAreTwoRowsOfEight(t *testing.T) {
	text, err := written(dusk(t, "dusk"), "h")
	if err != nil {
		t.Fatal(err)
	}
	row := regexp.MustCompile(`^  ("#[0-9a-f]{6}", ){7}"#[0-9a-f]{6}",$`)
	var rows int
	for line := range strings.SplitSeq(text, "\n") {
		if row.MatchString(line) {
			rows++
		}
		if strings.HasPrefix(line, `  "#`) && !row.MatchString(line) {
			t.Errorf("a row that is not eight colours: %q", line)
		}
	}
	if rows != rowsOfEight {
		t.Errorf("%d rows of eight, want %d", rows, rowsOfEight)
	}
}
