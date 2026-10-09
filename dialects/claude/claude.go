// Package claude is the claude code dialect: the colours claude code
// draws its own screen in, by name.
//
// Claude code takes no theme file. When the theme follows the terminal,
// it draws in numbered palette colours. Its text is 252. The words typed
// to it are 188, on a band of 0.
//
// So its colours are set in the terminal theme's palette. A colourway
// names them by what they are. This dialect is that table, and it moves
// between the names and the numbers.
//
// The band is palette 0, which is also the terminal's black. A colourway
// sets it as black. The dialect does not keep a second name for it,
// because two names could disagree.
package claude

import (
	"fmt"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// Slots are claude code's colours. Each has the role a colourway names it
// by, the palette number claude code draws it in, and what it is for.
//
// From is the general role a slot follows when a colourway does not set
// it. Toward is used for the bands of a diff. It says how far the ground
// is mixed toward that role, instead of taking the role's colour.
var Slots = []struct {
	Role   string
	Number int
	Use    string
	From   string
	Toward float64
}{
	{"claude-text", 252, "its text, the prompt mark and the status line",
		"ink", 0},
	{"claude-words", 188, "the words typed to it, and the dot of a reply",
		"ink", 0},
	{"claude-code", 153, "a code span in a reply", "code", 0},
	{"black", 0, "the band behind those words; the terminal's black",
		"", 0},
	{"claude-mode", 180, "the mode line", "info", 0},
	{"claude-rules", 175, "the rules round the prompt, and the badge",
		"border", 0},
	{"claude-spinner", 219, "the spinner, its dot, and the input cursor",
		"heading", 0},
	{"claude-added", 22, "the band of an added line in a diff", "ok", 0.25},
	{"claude-removed", 52, "the band of a removed line in a diff",
		"err", 0.25},
	{"claude-diff-code", 231, "the code on those bands", "ink", 0},
}

// Palette is the claude code colours a colourway sets, keyed by palette
// number. A terminal or a multiplexer is told these numbers.
func Palette(roles *css.Sheet) map[int]swatch.Swatch {
	palette := map[int]swatch.Swatch{}
	for _, slot := range Slots {
		if value, found := roles.Get(slot.Role); found {
			palette[slot.Number] = value
		}
	}
	return palette
}

// Into is a colourway with its claude code colours also set by number, as
// palette-N. The ghostty dialect then writes them into the terminal
// theme. Slots below 16 are skipped. If a colourway sets both a name and
// its number, Into returns an error.
func Into(roles *css.Sheet) (*css.Sheet, error) {
	said := css.New()
	for _, rule := range roles.Rules() {
		said.Set(rule.Name, rule.Swatch)
	}
	for _, slot := range Slots {
		value, found := roles.Get(slot.Role)
		if !found || slot.Number < 16 {
			continue
		}
		number := fmt.Sprintf("palette-%d", slot.Number)
		if _, both := roles.Get(number); both {
			return nil, fmt.Errorf(
				"claude: %s and %s are one colour, set twice",
				slot.Role, number)
		}
		said.Set(number, value)
	}
	return said, nil
}

// Of is a terminal palette's claude code colours as a colourway's roles,
// for the numbers the palette has. The band is left out. Below 16 the
// terminal already names the colour, black. The caller should drop
// palette-N roles for the same numbers, since these roles name them.
func Of(palette map[int]swatch.Swatch) *css.Sheet {
	roles := css.New()
	for _, slot := range Slots {
		value, found := palette[slot.Number]
		if found && slot.Number >= 16 {
			roles.Set(slot.Role, value)
		}
	}
	return roles
}
