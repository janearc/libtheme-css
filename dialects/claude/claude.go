// Package claude is the claude code dialect: the colours claude code
// draws its own screen in, by name.
//
// Claude code takes no theme file. With a theme that follows the
// terminal, it draws in numbered palette colours: its text in 252, the
// words typed to it in 188 on a band of 0, and so on.
//
// So its colours are set where the terminal's are, in the terminal
// theme's palette, and a colourway names them by what they are. This
// dialect is that table, read off claude code's screen on 2026-10-04,
// and the move between the names and the numbers.
//
// The band is palette 0, which is also the terminal's black: one lamp,
// two names. A colourway sets it as black, and the dialect says so
// rather than keeping a second name that could disagree.
package claude

import (
	"fmt"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// Slots are claude code's colours: the role a colourway names each by,
// the palette number claude code draws it in, and what it is.
//
// From is the general role a slot follows when a colourway does not set
// it, and Toward, for the bands of a diff, how far the ground is mixed
// toward that role rather than taking it. A first draft, 2026-10-05.
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

// Palette is the claude code colours a colourway sets, by palette
// number, which is what a terminal or a multiplexer is told.
func Palette(roles *css.Sheet) map[int]swatch.Swatch {
	palette := map[int]swatch.Swatch{}
	for _, slot := range Slots {
		if value, found := roles.Get(slot.Role); found {
			palette[slot.Number] = value
		}
	}
	return palette
}

// Into is a colourway with its claude code colours also said by number,
// as palette-N, so the ghostty dialect writes them into the terminal
// theme. A colourway that sets both a name and its number disagrees with
// itself, and is refused.
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
// for the numbers it has. The band is left out: below 16 a colour has
// the terminal's name already, black. palette-N roles for the same
// numbers should be dropped by the caller, since these name them.
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
