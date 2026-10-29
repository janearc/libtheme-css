package colourway_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/colourway"
	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// pathsOf are the files a render writes for a name, under the folder.
func pathsOf(name string) []string {
	return []string{filepath.Join("ghostty", name),
		filepath.Join("nvim", "colors", name+".lua"),
		filepath.Join("vim", "colors", name+".vim"),
		filepath.Join("glamour", name+".json")}
}

// hexes are every #rrggbb in a file, in order.
var hexes = regexp.MustCompile(`#[0-9a-fA-F]{6}`)

// all asks regexp for every match. It is defined by the regexp package.
const all = -1

// complete is what a byte and its complement add up to. It is defined by
// the byte.
const complete = 255

// opposite is whether two colours are each other's complement, channel by
// channel, on the byte grid.
func opposite(first, second string) bool {
	red, green, blue := bytesOf(first)
	redTwin, greenTwin, blueTwin := bytesOf(second)
	return red+redTwin == complete && green+greenTwin == complete &&
		blue+blueTwin == complete
}

// bytesOf are a hex colour's three bytes.
func bytesOf(hex string) (red, green, blue int) {
	lit := srgb.MustHex(hex)
	low, mid, high := lit.Bytes()
	return int(low), int(mid), int(high)
}

// twin is one file's colours, normal and inverted.
type twin struct {
	file             string
	normal, inverted []string
}

// colours are every #rrggbb in the file at a path under a folder.
func coloursIn(dir, path string) ([]string, error) {
	raw, err := os.ReadFile(filepath.Join(dir, path))
	return hexes.FindAllString(string(raw), all), err
}

// twinsOf renders a source and returns each file's colours, normal and
// inverted.
func twinsOf(source colourway.Source) ([]twin, error) {
	dir, err := os.MkdirTemp("", "invert-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	for _, made := range colourway.Render(source, dir) {
		if made.Err != nil {
			return nil, fmt.Errorf("%s for %s: %w", made.Program, made.Name, made.Err)
		}
	}
	var twins []twin
	inverted := pathsOf(source.Name + "-inverted")
	for place, file := range pathsOf(source.Name) {
		one := twin{file: file}
		if one.normal, err = coloursIn(dir, file); err != nil {
			return nil, err
		}
		if one.inverted, err = coloursIn(dir, inverted[place]); err != nil {
			return nil, err
		}
		twins = append(twins, one)
	}
	return twins, nil
}

// plan is the colours a source is made of. They are the ground and the ink
// and the sixteen, and any roles it sets itself. Roles not set are derived.
// The panel and the selection and the cursor line are among them.
type plan struct {
	ground, ink string
	sixteen     [len(ghostty.Ansi)]string
	// overrides are roles the source sets itself. They are by name. They
	// would otherwise be derived.
	overrides map[string]string
}

// sourceOf is the source a plan describes.
func sourceOf(colours plan) colourway.Source {
	roles := css.New()
	roles.Set("ground", srgb.MustHex(colours.ground).Swatch())
	roles.Set("ink", srgb.MustHex(colours.ink).Swatch())
	for number, name := range ghostty.Ansi {
		roles.Set(name, srgb.MustHex(colours.sixteen[number]).Swatch())
	}
	for name, colour := range colours.overrides {
		roles.Set(name, srgb.MustHex(colour).Swatch())
	}
	var made colourway.Source
	made.Name, made.About, made.Roles = "dusk", "a test", roles
	made.Origin = "sources/dusk.css"
	return made
}

// wrong lists every colour of every file whose inverted twin is not its
// complement.
func wrong(colours plan) ([]string, error) {
	twins, err := twinsOf(sourceOf(colours))
	var found []string
	for _, one := range twins {
		if len(one.normal) == nothing || len(one.normal) != len(one.inverted) {
			said := fmt.Sprintf("%s: %d colours and %d in the twin",
				one.file, len(one.normal), len(one.inverted))
			found = append(found, said)
			continue
		}
		for place, colour := range one.normal {
			if !opposite(colour, one.inverted[place]) {
				found = append(found, fmt.Sprintf("%s colour %d: %s and %s",
					one.file, place, colour, one.inverted[place]))
			}
		}
	}
	return found, err
}

// The finding's own case: a black ground, a white ink and no panel. Every
// colour the twin writes, the derived panel and selection and cursor line
// and claude code's band colours among them, is the complement of the
// normal file's, so a screen the system inverts shows what was meant.
func TestTheInvertedTwinIsTheComplementOfEveryColour(t *testing.T) {
	var colours plan
	colours.ground, colours.ink = "#000000", "#ffffff"
	for number := range colours.sixteen {
		colours.sixteen[number] = "#808080"
	}
	found, err := wrong(colours)
	if err != nil {
		t.Fatal(err)
	}
	for _, each := range found {
		t.Error(each)
	}
}

// overridden are the roles a source may set itself and not leave to be
// derived. They are the general vocabulary and claude code's slots. They
// also include a palette number, the cursor and neovim's override of the ink.
func overridden() []string {
	names := slices.Concat(colourway.Grounds, colourway.Text, colourway.States)
	return append(names, "claude-text", "claude-added", "palette-100",
		"cursor", "cursor-ink", "selection-ink", "nvim-ink")
}

// Roles the source sets itself are complemented as the derived ones are,
// across all four kinds of file. The case is every one of the names above
// set at once, each to a colour of its own.
func TestTheInvertedTwinComplementsEveryOverride(t *testing.T) {
	var colours plan
	colours.ground, colours.ink = "#101820", "#e8e0d0"
	for number := range colours.sixteen {
		colours.sixteen[number] = fmt.Sprintf("#%02x%02x%02x", greyStep*number,
			greyStep*number, greyStep*number)
	}
	colours.overrides = map[string]string{}
	for place, name := range overridden() {
		colours.overrides[name] = fmt.Sprintf("#%02x%02x%02x", greyStep*place,
			complete-greyStep*place, overrideBlue)
	}
	found, err := wrong(colours)
	if err != nil {
		t.Fatal(err)
	}
	for _, each := range found {
		t.Error(each)
	}
}

// greyStep is the gap between one grey and the next. It is defined by this
// test: it keeps twenty-four greys below 256.
const greyStep = 9

// overrideBlue is the blue of every override. It is defined by this test.
const overrideBlue = 77

// twinIsComplementWithOverrides is the property for a source that sets
// roles of its own: any colours for any of the names above.
func twinIsComplementWithOverrides(t *rapid.T) {
	colour := rapid.StringMatching(`#[0-9a-f]{6}`)
	var colours plan
	colours.ground, colours.ink = colour.Draw(t, "ground"), colour.Draw(t, "ink")
	for number := range colours.sixteen {
		colours.sixteen[number] = colour.Draw(t, "colour")
	}
	colours.overrides = map[string]string{}
	for _, name := range overridden() {
		if rapid.Bool().Draw(t, "set "+name) {
			colours.overrides[name] = colour.Draw(t, name)
		}
	}
	found, err := wrong(colours)
	if err != nil {
		t.Fatal(err)
	}
	for _, each := range found {
		t.Error(each)
	}
}

// The property for sources that set some of those roles and not others.
func TestTheInvertedTwinIsTheComplementWithAnyOverrides(t *testing.T) {
	rapid.Check(t, twinIsComplementWithOverrides)
}

// The property: for any ground, ink and sixteen, every colour of the
// inverted twin is the complement of the normal one.
func TestTheInvertedTwinIsTheComplementWhateverTheColours(t *testing.T) {
	rapid.Check(t, twinIsComplement)
}

// twinIsComplement is one draw of the property above.
func twinIsComplement(t *rapid.T) {
	colour := rapid.StringMatching(`#[0-9a-f]{6}`)
	var colours plan
	colours.ground, colours.ink = colour.Draw(t, "ground"), colour.Draw(t, "ink")
	for number := range colours.sixteen {
		colours.sixteen[number] = colour.Draw(t, "colour")
	}
	found, err := wrong(colours)
	if err != nil {
		t.Fatal(err)
	}
	for _, each := range found {
		t.Error(each)
	}
}
