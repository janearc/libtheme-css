package main

import (
	"flag"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// update rewrites the golden files from what the pages print now. Read the
// diff before you commit it.
var update = flag.Bool("update", false, "rewrite the golden files")

// goldenDir is where the pages' expected output is kept.
const goldenDir = "testdata/golden"

// readable is the permission of a golden file. It is defined by the shell.
const readable = 0o644

// folder is the permission of the golden folder. It is defined by the shell.
const folder = 0o755

// goldenFor is the expected text in a golden file, or the file's name and
// the text written there when update is set.
func goldenFor(t *testing.T, name, got string) string {
	t.Helper()
	path := filepath.Join(goldenDir, name)
	if *update {
		if err := os.MkdirAll(goldenDir, folder); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), readable); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(want)
}

// Each page paints what its golden file holds, escapes included. That holds
// the place, width and colour of every cell.
func TestPagesPaintTheirGolden(t *testing.T) {
	for _, each := range pages {
		got := capture(t, each.show)
		if want := goldenFor(t, each.name+".txt", got); got != want {
			t.Errorf("%s: painted\n%q\nwant\n%q", each.name, got, want)
		}
	}
}

// Each page says exactly its golden stylesheet.
func TestPagesSayTheirGoldenCSS(t *testing.T) {
	for _, each := range pages {
		got := each.css()
		if want := goldenFor(t, each.name+".css", got); got != want {
			t.Errorf("%s: css\n%q\nwant\n%q", each.name, got, want)
		}
	}
}

// The anti-test: the golden holds only its own page. Every page differs from
// every other, so a page that painted another's output would be caught.
func TestGoldensDifferFromEachOther(t *testing.T) {
	seen := map[string]string{}
	for _, each := range pages {
		got := capture(t, each.show)
		if other, again := seen[got]; again {
			t.Errorf("%s paints what %s paints", each.name, other)
		}
		seen[got] = each.name
	}
}

// The anti-test: a page painted with one byte changed is not its golden.
func TestAChangedPageIsNotItsGolden(t *testing.T) {
	for _, each := range pages {
		got := capture(t, each.show)
		changed := got[:len(got)-lastByte] + "!"
		if changed == goldenFor(t, each.name+".txt", got) {
			t.Errorf("%s: a changed page matched its golden", each.name)
		}
	}
}

// Black lifts to itself. It isn't divided by its own nothing. The next
// test is its anti-test: it lifts a colour with something in it.
func TestLiftingBlackLeavesBlack(t *testing.T) {
	black := swatch.NoLight()
	if got := lifted(black); got != black {
		t.Errorf("black lifted to %v", got)
	}
}

// Lifting a colour with something in it fills its brightest lamp.
func TestLiftingFillsTheBrightestLamp(t *testing.T) {
	dim := srgb.MustHex(dimHex)
	raised := lifted(dim.Swatch())
	back, _ := srgb.FromSwatch(raised)
	if math.Abs(back.G-fullLamp) > tinyLift ||
		math.Abs(back.R-dim.R/dim.G) > tinyLift {
		t.Errorf("lifted to %v", back)
	}
}

const (
	// dimHex is a colour whose brightest lamp is green, and is not full. It
	// is defined by this test.
	dimHex = "#1a330d"
	// fullLamp is a lamp at full. It is defined by srgb.
	fullLamp = 1
	// lastByte is how far from the end of a page the byte that is changed is.
	// It is defined by counting.
	lastByte = 1
)

// tinyLift is how far a lifted lamp may be from its value and be the same.
// It is defined by this test.
const tinyLift = 1e-6
