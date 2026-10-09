// Package ghostty is the ghostty dialect: a terminal theme said from a
// colourway's roles, and read back into them.
//
// Ghostty takes a theme as a file of settings. They are the ground, the
// ink, the cursor and the selection. The file also has a palette of 256
// numbered colours. Programs ask for the first sixteen by name.
//
// A colourway names each of these by role. This dialect is the table
// between the two, in both directions. It also knows the file's format,
// so nobody has to write a theme by hand.
//
// The roles are the ones the sheets already used: ground, ink, cursor,
// cursor-ink, surface-1 for the selection, selection-ink and the sixteen
// by name. palette-N sets any entry from 16 to 255. A sheet uses it to
// reach a colour a program picks by number.
package ghostty

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// Keys are ghostty's colour settings in the order a theme writes them,
// each with the role a colourway names it by.
var Keys = []struct{ Key, Role string }{
	{"background", "ground"},
	{"foreground", "ink"},
	{"cursor-color", "cursor"},
	{"cursor-text", "cursor-ink"},
	{"selection-background", "surface-1"},
	{"selection-foreground", "selection-ink"},
}

// Ansi are the sixteen colours a terminal names, by palette number.
var Ansi = [16]string{
	"black", "red", "green", "yellow",
	"blue", "magenta", "cyan", "white",
	"bright-black", "bright-red", "bright-green", "bright-yellow",
	"bright-blue", "bright-magenta", "bright-cyan", "bright-white",
}

// Theme is a ghostty theme's colours. Settings is keyed by setting name.
// Palette is keyed by palette number.
type Theme struct {
	Settings map[string]swatch.Swatch
	Palette  map[int]swatch.Swatch
}

// empty is a theme with nothing in it yet.
func empty() Theme {
	return Theme{Settings: map[string]swatch.Swatch{},
		Palette: map[int]swatch.Swatch{}}
}

// Of is the theme a colourway's roles describe. The ground, the ink and
// the sixteen are required. A terminal missing any of them would draw in
// someone else's colours. The other settings are written when the
// colourway has them.
func Of(roles *css.Sheet) (Theme, error) {
	theme := empty()
	for _, setting := range Keys {
		value, found := roles.Get(setting.Role)
		if !found {
			if setting.Role == "ground" || setting.Role == "ink" {
				return Theme{}, missing(setting.Role)
			}
			continue
		}
		theme.Settings[setting.Key] = value
	}
	for number, name := range Ansi {
		value, found := roles.Get(name)
		if !found {
			return Theme{}, missing(name)
		}
		theme.Palette[number] = value
	}
	for _, name := range roles.Names() {
		digits, found := strings.CutPrefix(name, "palette-")
		if !found {
			continue
		}
		number, err := strconv.Atoi(digits)
		if err != nil || number < 16 || number > 255 {
			return Theme{}, fmt.Errorf("ghostty: %s: an entry "+
				"is 16 to 255; below, use its name", name)
		}
		theme.Palette[number], _ = roles.Get(name)
	}
	return theme, nil
}

// Roles is the theme said back as a colourway's roles. The settings and
// the sixteen are named by role. Every other palette entry becomes
// palette-N. They come in the order a theme writes them.
func (theme Theme) Roles() *css.Sheet {
	roles := css.New()
	for _, setting := range Keys {
		if value, found := theme.Settings[setting.Key]; found {
			roles.Set(setting.Role, value)
		}
	}
	for number, name := range Ansi {
		if value, found := theme.Palette[number]; found {
			roles.Set(name, value)
		}
	}
	for _, number := range theme.above16() {
		name := fmt.Sprintf("palette-%d", number)
		roles.Set(name, theme.Palette[number])
	}
	return roles
}

// above16 are the palette's numbers past the sixteen, in order.
func (theme Theme) above16() []int {
	numbers := []int{}
	for number := range theme.Palette {
		if number >= 16 {
			numbers = append(numbers, number)
		}
	}
	sort.Ints(numbers)
	return numbers
}

// cellColour are ghostty's own words for a colour. A setting may use one
// in place of a hex. They are not colours a colourway can hold, so Read
// skips them.
var cellColour = map[string]bool{"cell-foreground": true,
	"cell-background": true}

// Read is a theme file's colours. It skips comments, blank lines and
// settings that are not colours. A colour that does not parse is an error
// that names its line.
func Read(source io.Reader) (Theme, error) {
	theme := empty()
	colours := map[string]bool{}
	for _, setting := range Keys {
		colours[setting.Key] = true
	}
	scanner := bufio.NewScanner(source)
	for number := 1; scanner.Scan(); number++ {
		text := strings.TrimSpace(scanner.Text())
		key, value, found := strings.Cut(text, "=")
		if text == "" || strings.HasPrefix(text, "#") || !found {
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		switch {
		case key == "palette":
			index, hex, found := strings.Cut(value, "=")
			entry, err := strconv.Atoi(strings.TrimSpace(index))
			if !found || err != nil || entry < 0 || entry > 255 {
				return Theme{}, atLine(number,
					fmt.Errorf("palette %q", value))
			}
			if theme.Palette[entry], err = parse(hex); err != nil {
				return Theme{}, atLine(number, err)
			}
		case colours[key] && cellColour[value]:
			continue
		case colours[key]:
			var err error
			if theme.Settings[key], err = parse(value); err != nil {
				return Theme{}, atLine(number, err)
			}
		}
	}
	return theme, scanner.Err()
}

// Write puts the theme in ghostty's format after a header. Each line of
// the header becomes a comment. The settings come first, then the
// sixteen, then any entries above them. Only what the theme has is
// written. A theme read from a partial file is written back partial.
func (theme Theme) Write(out io.Writer, header string) error {
	var text strings.Builder
	for line := range strings.SplitSeq(strings.TrimSpace(header), "\n") {
		fmt.Fprintln(&text, strings.TrimRight("# "+line, " "))
	}
	fmt.Fprintln(&text)
	for _, setting := range Keys {
		if value, found := theme.Settings[setting.Key]; found {
			fmt.Fprintf(&text, "%s = %s\n", setting.Key, Hex(value))
		}
	}
	fmt.Fprintln(&text)
	for number := range 16 {
		if value, found := theme.Palette[number]; found {
			fmt.Fprintf(&text, "palette = %d=%s\n", number,
				Hex(value))
		}
	}
	high := theme.above16()
	if len(high) > 0 {
		fmt.Fprintln(&text)
	}
	for _, number := range high {
		fmt.Fprintf(&text, "palette = %d=%s\n", number,
			Hex(theme.Palette[number]))
	}
	_, err := io.WriteString(out, text.String())
	return err
}

// Inverted is the theme with every colour turned to its opposite. Each
// lamp becomes 255 less itself. A screen that the operating system
// inverts then draws it as the original.
func (theme Theme) Inverted() Theme {
	inverted := empty()
	for key, value := range theme.Settings {
		inverted.Settings[key] = invert(value)
	}
	for number, value := range theme.Palette {
		inverted.Palette[number] = invert(value)
	}
	return inverted
}

// missing is the error for a role a theme cannot do without.
func missing(role string) error {
	return fmt.Errorf("ghostty: %s is missing", role)
}

// atLine is an error with the line of the file it was found on.
func atLine(number int, err error) error {
	return fmt.Errorf("ghostty: line %d: %w", number, err)
}

// Hex is a swatch as the #rrggbb a theme file holds. It is the nearest
// colour the lamps can make.
func Hex(value swatch.Swatch) string {
	lamps, _ := srgb.FromSwatch(value)
	return lamps.Hex()
}

// parse reads a #rrggbb.
func parse(hex string) (swatch.Swatch, error) {
	lamps, err := srgb.FromHex(strings.TrimSpace(hex))
	if err != nil {
		return swatch.Swatch{}, err
	}
	return lamps.Swatch(), nil
}

// invert is a colour's opposite on the lamps.
func invert(value swatch.Swatch) swatch.Swatch {
	lamps, _ := srgb.FromSwatch(value)
	r, g, b := lamps.Bytes()
	return srgb.RGB8(255-r, 255-g, 255-b).Swatch()
}
